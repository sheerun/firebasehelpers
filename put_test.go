package firebasehelpers

import (
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
)

func PutString(source string, path string, data string) string {
	var sourceInterface interface{}
	var dataInterface interface{}
	json.Unmarshal([]byte(source), &sourceInterface)
	json.Unmarshal([]byte(data), &dataInterface)
	resultInterface := Put(sourceInterface, path, dataInterface)
	result, err := json.Marshal(resultInterface)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func TestPutAtRoot(t *testing.T) {
	source := `{"foo":"bar"}`
	target := `{"fiz":"fuz"}`
	result := PutString(source, "/", target)

	assert.Equal(t, target, result)
}

func TestPutAtPath(t *testing.T) {
	source := `{"foo":"bar"}`
	target := `{"foo":"fuz"}`
	result := PutString(source, "/foo", `"fuz"`)

	assert.Equal(t, target, result)
}

func TestAddAtPath(t *testing.T) {
	source := `{"foo":"bar"}`
	target := `{"fiz":"fuz","foo":"bar"}`
	result := PutString(source, "/fiz", `"fuz"`)

	assert.Equal(t, target, result)
}

func TestReplaceNull(t *testing.T) {
	source := `null`
	target := `{"fiz":"fuz"}`
	result := PutString(source, "/fiz", `"fuz"`)

	assert.Equal(t, target, result)
}

func TestReplaceString(t *testing.T) {
	source := `"hello"`
	target := `{"fiz":"fuz"}`
	result := PutString(source, "/fiz", `"fuz"`)

	assert.Equal(t, target, result)
}

func TestSetNested(t *testing.T) {
	source := `{"foo":{"bar":"baz"}}`
	target := `{"foo":{"bar":"buz"}}`
	result := PutString(source, "/foo/bar", `"buz"`)

	assert.Equal(t, target, result)
}

func TestReplaceStringNested(t *testing.T) {
	source := `"foo"`
	target := `{"foo":{"bar":"buz"}}`
	result := PutString(source, "/foo/bar", `"buz"`)

	assert.Equal(t, target, result)
}

func TestAddNested(t *testing.T) {
	source := `{"foo":{"bar":"buz"}}`
	target := `{"foo":{"bar":"buz","fiz":"fuz"}}`
	result := PutString(source, "/foo/fiz", `"fuz"`)

	assert.Equal(t, target, result)
}

func TestRemoveNested(t *testing.T) {
	source := `{"foo":{"bar":"buz"}}`
	target := `null`
	result := PutString(source, "/foo/bar", `null`)

	assert.Equal(t, target, result)
}

func TestRemoveSeminested(t *testing.T) {
	source := `{"foo":{"fiz":"fuz","bar":"buz"}}`
	target := `{"foo":{"fiz":"fuz"}}`
	result := PutString(source, "/foo/bar", `null`)

	assert.Equal(t, target, result)
}

func TestRemoveDummy(t *testing.T) {
	source := `{"fiz":"fuz","foo":{"bar":"buz"}}`
	target := `{"fiz":"fuz","foo":{"bar":"buz"}}`
	result := PutString(source, "/foo/fuz/lol", `null`)

	assert.Equal(t, target, result)
}

func TestComplex(t *testing.T) {
	source := `null`
	target := `{"fiz":"fuz","foo":{"bar":"buz"}}`
	source = PutString(source, "/foo", `{"bar":"biz"}`)
	source = PutString(source, "/foo/bar", `"buz"`)
	source = PutString(source, "/fiz", `"fuz"`)
	assert.Equal(t, target, source)
}
