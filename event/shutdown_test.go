package event

import (
	"testing"

	"github.com/xytis/gami"
)

func TestShutdownEvent(t *testing.T) {
	fixture := map[string]string{
		"Shutdown": "Shutdown",
		"Restart":  "Restart",
	}

	ev := gami.AMIEvent{
		ID:        "Shutdown",
		Privilege: []string{"all"},
		Params:    fixture,
	}

	evtype := New(&ev)
	if _, ok := evtype.(Shutdown); !ok {
		t.Log("Shutdown type assertion")
		t.Fail()
	}

	testEvent(t, fixture, evtype)
}
