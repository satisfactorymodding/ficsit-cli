package utils

import (
	"bytes"
	"fmt"
	"time"
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
		return fmt.Errorf("failed to parse date time: %w", err)
	}

	*v = parsed

	return nil
}
