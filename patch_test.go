package firebasehelpers

import (
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
)

func PatchString(source string, path string, data string) string {
	var sourceInterface interface{}
	var dataInterface interface{}
	json.Unmarshal([]byte(source), &sourceInterface)
	json.Unmarshal([]byte(data), &dataInterface)
	resultInterface := Patch(sourceInterface, path, dataInterface)
	result, err := json.Marshal(resultInterface)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func TestPatch(t *testing.T) {
	source := `{"fiz":"fuz","foo":{"bar":"buz"}}`
	target := `{"fiz":"fuz","foo":{"fiz":"fuz"}}`
	source = PatchString(source, "/foo", `{"bar":null,"fiz":"fuz"}`)
	assert.Equal(t, target, source)
}

func TestPatchShallow(t *testing.T) {
	source := `{"fiz":"fuz","foo":{"bar":"buz"}}`
	target := `{"fiz":"faz","foo":{"foo":"bar"}}`
	source = PatchString(source, "/", `{"fiz":"faz","foo":{"foo":"bar"}}`)
	assert.Equal(t, target, source)
}
