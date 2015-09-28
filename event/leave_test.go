// Package event for AMI
package event

import (
	"testing"

	"github.com/xytis/gami"
)

func TestLeave(t *testing.T) {
	fixture := map[string]string{
		"Queue":    "Queue",
		"Count":    "Count",
		"Position": "Position",
		"Channel":  "Channel",
		"Uniqueid": "UniqueID",
	}

	ev := gami.AMIEvent{
		ID:        "Leave",
		Privilege: []string{"all"},
		Params:    fixture,
	}

	evtype := New(&ev)
	if _, ok := evtype.(Leave); !ok {
		t.Fatal("Leave type assertion")
	}

	testEvent(t, fixture, evtype)
}
