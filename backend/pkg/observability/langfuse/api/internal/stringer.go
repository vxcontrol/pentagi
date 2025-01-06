package internal

import "encoding/json"

// StringifyJSON returns a pretty JSON string representation of
// the given value.
func StringifyJSON(value interface{}) (string, error) {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
