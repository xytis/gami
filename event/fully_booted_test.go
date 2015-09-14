package event

import (
	"testing"

	"github.com/xytis/gami"
)

func TestFullyBottedEvent(t *testing.T) {
	fixture := map[string]string{
		"Status": "Status",
	}

	ev := gami.AMIEvent{
		ID:        "FullyBooted",
		Privilege: []string{"all"},
		Params:    fixture,
	}

	evtype := New(&ev)
	if _, ok := evtype.(FullyBooted); !ok {
		t.Log("FullyBooted type assertion")
		t.Fail()
	}

	testEvent(t, fixture, evtype)
}
