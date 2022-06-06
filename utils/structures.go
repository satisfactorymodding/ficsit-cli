package utils

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func Copy[T any](obj T) (*T, error) {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal object")
	}

	out := new(T)
	if err := json.Unmarshal(marshal, out); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal object")
	}

	return out, nil
}
