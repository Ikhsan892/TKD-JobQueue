package utils

import "encoding/json"

func ParsePayload[T any](processName, param string) (T, error) {
	var payload T
	// parsing payload
	if err := json.Unmarshal([]byte(param), &payload); err != nil {
		Error(processName, err)
		return payload, err
	}
	return payload, nil
}
