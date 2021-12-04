package utils

import (
	"bytes"
	"time"

	"github.com/pkg/errors"
)

//goland:noinspection GoUnusedExportedFunction
func UnmarshalDateTime(b []byte, v *time.Time) error {
	trimmed := bytes.Trim(b, "\"")

	if len(trimmed) == 0 {
		*v = time.Unix(0, 0)
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, string(trimmed))
	if err != nil {
		return errors.Wrap(err, "failed to parse date time")
	}

	*v = parsed

	return nil
}
