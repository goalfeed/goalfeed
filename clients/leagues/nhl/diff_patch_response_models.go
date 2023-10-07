package nhl

import (
	"encoding/json"
)

type NHLDiffSet struct {
	Diff []NHLDiffItem `json:"diff"`
}
type NHLDiffItem struct {
	Path  string          `json:"path"`
	Op    string          `json:"op"`
	Value json.RawMessage `json:"value"`
}
type NHLDiffPatch []NHLDiffSet
