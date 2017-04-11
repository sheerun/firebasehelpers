package firebasehelpers

import "strings"

func patch(object interface{}, keys []string, value map[string]interface{}) interface{} {
	for key, value := range value {
		object = put(object, append(keys, key), value)
	}

	return object
}

func Patch(object interface{}, path string, value interface{}) interface{} {
	switch value := value.(type) {
	case map[string]interface{}:
		keys := strings.Split(path[1:], "/")

		// If path is "/" then keys is []string{""}, not []string{}
		if keys[0] == "" {
			keys = []string{}
		}

		return patch(object, keys, value)
	}

	return object
}
