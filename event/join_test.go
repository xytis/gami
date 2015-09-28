// Package event for AMI
package event

import (
	"testing"

	"github.com/xytis/gami"
)

func TestJoin(t *testing.T) {
	fixture := map[string]string{
		"Queue":             "Queue",
		"Position":          "Position",
		"Count":             "Count",
		"Channel":           "Channel",
		"CallerIDNum":       "CallerIDNum",
		"CallerIDName":      "CallerIDName",
		"ConnectedLineNum":  "ConnectedLineNum",
		"ConnectedLineName": "ConnectedLineName",
		"Uniqueid":          "UniqueID",
	}

	ev := gami.AMIEvent{
		ID:        "Join",
		Privilege: []string{"all"},
		Params:    fixture,
	}

	evtype := New(&ev)
	if _, ok := evtype.(Join); !ok {
		t.Fatal("Join type assertion")
	}

	testEvent(t, fixture, evtype)
}
