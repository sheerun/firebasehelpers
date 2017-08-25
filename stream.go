package firebasehelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	j "github.com/bitly/go-simplejson"
	"github.com/cenkalti/backoff"
	"github.com/desertbit/timer"
	"github.com/knq/firebase"
	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
	"github.com/sheerun/yson"
)

type event struct {
	Path []string
	Prev []byte
	Curr []byte
}

type Stream struct {
	errHandler   func(error)
	value        []byte
	in           chan []byte
	ShutdownChan chan struct{}
	stop         chan struct{}
	stopped      bool
	listeners    []*Listener
	mux          sync.Mutex
	processMux   sync.Mutex
	shutdownMux  sync.Mutex
	wg           sync.WaitGroup
}

type Listener struct {
	cursor *cursor
	value  []byte
	cb     func(path []string, prev []byte, curr []byte)
	mux    sync.Mutex
}

func matches(json []byte, pattern []string) [][]string {
	keys := []string{}

	if len(pattern) > 0 {
		if pattern[0] == "*" {
			yson.EachKey(json, func(key []byte) {
				keys = append(keys, string(key))
			})
		} else {
			if yson.Get(json, pattern[0]) != nil {
				keys = append(keys, string(pattern[0]))
			}
		}

		if len(pattern) > 1 {
			result := [][]string{}

			for _, key := range keys {
				for _, path := range matches(yson.Get(json, key), pattern[1:]) {
					result = append(result, append([]string{key}, path...))
				}
			}

			return result
		} else {
			result := [][]string{}
			for _, key := range keys {
				result = append(result, []string{key})
			}
			return result
		}
	}

	return [][]string{}
}

func has(paths [][]string, path []string) bool {
	search := strings.Join(path, ".")

	for _, path := range paths {
		if strings.Join(path, ".") == search {
			return true
		}
	}

	return false
}

func (w *Listener) shutdown() {
	w.cursor.stream.removeListen(w)
}

func (w *Listener) call(event event) {
	w.cb(event.Path, event.Prev, event.Curr)
}

func reverse(ss [][]string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func (w *Listener) processRemove(value []byte) {
	previous := w.value

	prevMatches := matches(previous, w.cursor.path)
	currMatches := matches(value, w.cursor.path)

	reverse(prevMatches)
	for _, prevMatch := range prevMatches {
		if !has(currMatches, prevMatch) {
			w.call(event{Path: prevMatch, Prev: yson.Get(previous, prevMatch...)})
		}
	}
}

func (w *Listener) processChange(value []byte) {
	previous := w.value

	currMatches := matches(value, w.cursor.path)

	for _, currMatch := range currMatches {
		prev := yson.Get(previous, currMatch...)
		curr := yson.Get(value, currMatch...)

		if bytes.Compare(prev, curr) != 0 {
			w.call(event{Path: currMatch, Prev: prev, Curr: curr})
		}

	}

	w.value = value
}

func (w *Stream) Push(value []byte) {
	w.mux.Lock()
	defer w.mux.Unlock()

Loop:
	for {
		select {
		case w.in <- value:
			break Loop
		default:
		}
	}
}

func keys(json []byte) map[string]struct{} {
	set := map[string]struct{}{}

	yson.EachKey(json, func(key []byte) {
		set[string(key)] = struct{}{}
	})

	return set
}

func (w *Stream) refresh() {
	w.Push(w.value)
}

func (w *Stream) process() {
	for {
		select {
		case value := <-w.in:
			w.processSingle(value)
		case <-w.stop:
			for {
				_, ok := <-w.in

				if !ok {
					break
				}
			}
			return
		}
	}
}

func (w *Stream) processSingle(value []byte) {
	w.processMux.Lock()
	defer w.processMux.Unlock()

	w.value = value

	for i := len(w.listeners) - 1; i >= 0; i-- {
		w.listeners[i].processRemove(value)
	}

	for i := 0; i < len(w.listeners); i++ {
		w.listeners[i].processChange(value)
	}
}

func (w *Stream) pubError(err error) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.errHandler(err)
	}()
}

// Shuts down stream and all its descendants
func (w *Stream) Shutdown() {
	w.shutdownMux.Lock()
	defer w.shutdownMux.Unlock()

	if w.stopped {
		return
	}

	w.stopped = true

	close(w.ShutdownChan)

	w.mux.Lock()
	defer w.mux.Unlock()

	for i := len(w.listeners) - 1; i >= 0; i-- {
		w.listeners[i].shutdown()
	}

	w.wg.Wait()
}

func (s *Stream) publishFile(path string) {
	contents, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}

	buffer := new(bytes.Buffer)

	err = json.Compact(buffer, contents)

	if err != nil {
		panic(err)
	}

	s.Push(buffer.Bytes())
}

func (w *Stream) Async(fn func(), label string) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		fn()
	}()
}

func (w *Stream) WatchFile(path string) *Stream {
	var err error

	go w.process()

	w.publishFile(path)

	wtch := watcher.New()

	wtch.SetMaxEvents(1)

	w.Async(func() {
		for {
			select {
			case <-wtch.Event:
				w.publishFile(path)
			case <-w.ShutdownChan:
				wtch.Close()
				return
			}
		}
	}, "filewatch")

	err = wtch.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	w.Async(func() {
		if err := wtch.Start(time.Second); err != nil {
			log.Fatalln(err)
		}
	}, "watcherstart")

	return w
}

func NewStream(errHandler func(error)) *Stream {
	return &Stream{
		errHandler:   errHandler,
		listeners:    []*Listener{},
		ShutdownChan: make(chan struct{}),
		stop:         make(chan struct{}),
		in:           make(chan []byte, 1),
	}
}

func (w *Stream) WatchFirebase(r *firebase.DatabaseRef) *Stream {
	w.Async(w.process, "process")

	// By default the value is nil, but we don't send it through channel
	var js interface{}

	operation := func() (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				switch r := rec.(type) {
				case error:
					err = r
				default:
					err = errors.New(fmt.Sprintf("%s", err))
				}
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel()

		evs, err := r.Watch(ctx)

		if err != nil {
			return errors.Wrap(err, "failed to watch")
		}

		t := timer.NewTimer(time.Second * 60)

		for {
			select {
			case e := <-evs:
				if e == nil {
					return errors.New("streaming ended")
				}

				if e.Type == firebase.EventTypeCancel {
					return errors.New("streaming cancelled")
				}

				if e.Type == firebase.EventTypeClosed {
					return errors.New("streaming closed")
				}

				if e.Type == firebase.EventTypeAuthRevoked {
					return errors.New("streaming auth revoked")
				}

				// Not only for firebase.EventTypeKeepAlive
				// as other keep-alives don't arrive if other values do
				// One can expect at least one event every 30 seconds
				t.Reset(time.Second * 40)

				if e.Type == firebase.EventTypePut || e.Type == firebase.EventTypePatch {
					payload, err := j.NewJson(e.Data)

					if err != nil {
						// We don't return an error because we don't need to re-esablish link
						w.pubError(errors.Wrap(err, "failed to parse event data"))
						break
					}

					path, err := payload.GetPath("path").String()

					if err != nil {
						w.pubError(errors.Wrap(err, "failed to parse event path"))
						break
					}

					data := payload.Get("data").Interface()

					if e.Type == firebase.EventTypePut {
						js = Put(js, path, data)
					} else {
						js = Patch(js, path, data)
					}

					str, err := json.Marshal(js)

					if err != nil {
						w.pubError(errors.Wrap(err, "failed to marshal json"))
						break
					}

					w.Push(str)
				}
			case <-t.C:
				return errors.New("failed to receive keep-alive signal")
			case <-w.ShutdownChan:
				return nil
			}
		}
	}

	notify := func(err error, next time.Duration) {
		w.pubError(err)
	}

	w.Async(func() {
		bf := backoff.NewExponentialBackOff()
		bf.MaxElapsedTime = 0
		err := backoff.RetryNotify(operation, bf, notify)

		if err != nil {
			// Should not ever happen with this MaxElapsedTime
			panic(err)
		}
	}, "backoff")

	return w
}

func (w *Stream) Select(path ...string) *cursor {
	return &cursor{stream: w, path: path}
}

func (c *cursor) Select(path ...string) *cursor {
	return &cursor{stream: c.stream, path: append(c.path, path...)}
}

type cursor struct {
	stream *Stream
	path   []string
}

func (c *Stream) Listen(pattern []string, cb func(path []string, prev []byte, curr []byte)) *Listener {
	return c.listen(c.Select(pattern...), cb)
}

func remove(slice []*Listener, s int) []*Listener {
	return append(slice[:s], slice[s+1:]...)
}

func (w *Stream) removeListen(listener *Listener) bool {
	w.processMux.Lock()
	defer w.processMux.Unlock()

	for i, list := range w.listeners {
		if list == listener {
			w.listeners = remove(w.listeners, i)
			listener.processRemove([]byte("{}"))
			return true
		}
	}

	return false
}

func (w *Stream) listen(cursor *cursor, cb func(path []string, prev []byte, curr []byte)) *Listener {
	w.processMux.Lock()
	defer w.processMux.Unlock()

	listener := &Listener{
		cb:     cb,
		cursor: cursor,
	}

	w.listeners = append(w.listeners, listener)

	w.Async(w.refresh, "refresh")

	return listener
}

func (w *cursor) Value() []byte {
	return yson.Get(w.stream.value, w.path...)
}
