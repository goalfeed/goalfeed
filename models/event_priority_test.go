package models

import "testing"

func TestEventPriorityString(t *testing.T) {
	cases := []struct {
		p EventPriority
		s string
	}{
		{PriorityLow, "low"},
		{PriorityNormal, "normal"},
		{PriorityHigh, "high"},
		{PriorityCritical, "critical"},
		{EventPriority(999), "normal"},
	}
	for _, c := range cases {
		if c.p.String() != c.s {
			t.Fatalf("expected %s, got %s", c.s, c.p.String())
		}
	}
}
