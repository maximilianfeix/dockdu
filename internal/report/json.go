package report

import (
	"encoding/json"
	"io"
)

// WriteJSON renders the report as indented JSON. Lists serialise as [] and
// never as null so consumers can iterate without nil checks.
func WriteJSON(w io.Writer, r *Report) error {
	if r.Warnings == nil {
		r.Warnings = []string{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
