package firebasehelpers

import "strings"

func emptyWithValue(keys []string, value interface{}) interface{} {
	if len(keys) == 0 {
		return value
	}

	if value == nil {
		return nil
	}

	return map[string]interface{}{keys[0]: emptyWithValue(keys[1:], value)}
}

func put(object interface{}, keys []string, value interface{}) interface{} {
	if len(keys) == 0 {
		return value
	}

	key := keys[0]

	switch typed := object.(type) {
	case map[string]interface{}:
		typed[key] = put(typed[key], keys[1:], value)

		if typed[key] == nil {
			delete(typed, key)
		}

		if len(typed) == 0 {
			return nil
		}

		return object
	}

	return emptyWithValue(keys, value)
}

func Put(object interface{}, path string, value interface{}) interface{} {
	keys := strings.Split(path[1:], "/")

	// If path is "/" then keys is []string{""}, not []string{}
	if keys[0] == "" {
		keys = []string{}
	}

	return put(object, keys, value)
}
