package utils

import (
	"encoding/json"
	"fmt"
)

func Copy[T any](obj T) (*T, error) {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal object: %w", err)
	}

	out := new(T)
	if err := json.Unmarshal(marshal, out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object: %w", err)
	}

	return out, nil
}
