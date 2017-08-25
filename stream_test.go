package firebasehelpers

import (
	"testing"

	"github.com/go-test/deep"
)

func TestMatches(t *testing.T) {
	json := []byte(`
		{
			"managers": {
				"foo": {
					"supervisors": {
						"fiz": "fiz",
						"fuz": "fuz"
					}
				},
				"bar": {
					"supervisors": {
						"no": "no",
						"way": "way"
					}
				},
				"fiz": {
					"fuz": "asdfa"
				}
			}
	`)

	pattern := []string{"managers", "*", "supervisors", "*"}

	result := matches(json, pattern)

	expected := [][]string{
		[]string{"managers", "foo", "supervisors", "fiz"},
		[]string{"managers", "foo", "supervisors", "fuz"},
		[]string{"managers", "bar", "supervisors", "no"},
		[]string{"managers", "bar", "supervisors", "way"},
	}

	if diff := deep.Equal(expected, result); diff != nil {
		t.Error(diff)
	}
}

func TestMatches2(t *testing.T) {
	json := []byte(`
		{
			"managers": {
				"foo": {
					"supervisors": {
						"fiz": "fiz",
						"fuz": "fuz"
					}
				},
				"bar": {
					"supervisors": {
						"no": "no",
						"way": "way"
					}
				},
				"fiz": {
					"fuz": "asdfa"
				}
			}
	`)

	pattern := []string{"managers", "*", "supervisors"}

	result := matches(json, pattern)

	expected := [][]string{
		[]string{"managers", "foo", "supervisors"},
		[]string{"managers", "bar", "supervisors"},
	}

	if diff := deep.Equal(expected, result); diff != nil {
		t.Error(diff)
	}
}

func TestMatchesEmpty(t *testing.T) {
	json := []byte(`{} `)

	pattern := []string{"managers", "*", "supervisors"}

	result := matches(json, pattern)

	expected := [][]string{}

	if diff := deep.Equal(expected, result); diff != nil {
		t.Error(diff)
	}
}

func TestMatchesEmpty2(t *testing.T) {
	json := []byte(`{"managers":"foobar"}`)

	pattern := []string{"managers"}

	result := matches(json, pattern)

	expected := [][]string{
		[]string{"managers"},
	}

	if diff := deep.Equal(expected, result); diff != nil {
		t.Error(diff)
	}
}

func TestMatchesEmpty3(t *testing.T) {
	json := []byte(`{"managers":"foobar"}`)

	pattern := []string{}

	result := matches(json, pattern)

	expected := [][]string{}

	if diff := deep.Equal(expected, result); diff != nil {
		t.Error(diff)
	}
}

func TestHas(t *testing.T) {
	if !has(
		[][]string{
			[]string{"foo", "bar", "baz"},
			[]string{"fiz", "fuz", "faz"},
		},
		[]string{"fiz", "fuz", "faz"},
	) {
		t.Error("Should match")
	}
}

func TestHasNot(t *testing.T) {
	if has(
		[][]string{
			[]string{"foo", "bar", "baz"},
			[]string{"fiz", "fuz", "faz"},
		},
		[]string{"fiz", "fuz", "www"},
	) {
		t.Error("Should match")
	}
}
