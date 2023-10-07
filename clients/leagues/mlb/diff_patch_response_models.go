package mlb

import (
	"encoding/json"
)

type MLBDiffSet struct {
	Diff []MLBDiffItem `json:"diff"`
}
type MLBDiffItem struct {
	Path  string          `json:"path"`
	Op    string          `json:"op"`
	Value json.RawMessage `json:"value"`
}
type MLBDiffPatch []MLBDiffSet
