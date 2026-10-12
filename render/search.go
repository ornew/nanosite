package render

import (
	"encoding/json"
	"github.com/ornew/nanosite/publish"
)

// WriteIndex lets a theme select its search schema and script prefix while
// sharing deterministic encoding and output writing. The payload must contain
// only published/searchable pages selected by the caller.
func WriteIndex(file string, payload any, prefix string) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if prefix != "" {
		b = append(append([]byte(prefix), b...), ';', '\n')
	}
	return publish.WriteFile(file, b)
}
