package event

import (
	"testing"

	"github.com/xytis/gami"
)

func TestMasqueradeEvent(t *testing.T) {
	fixture := map[string]string{
		"Clone":         "Clone",
		"CloneState":    "CloneState",
		"Original":      "Original",
		"OriginalState": "OriginalState",
	}

	ev := gami.AMIEvent{
		ID:        "Masquerade",
		Privilege: []string{"all"},
		Params:    fixture,
	}

	evtype := New(&ev)
	if _, ok := evtype.(Masquerade); !ok {
		t.Log("Masquerade type assertion")
		t.Fail()
	}

	testEvent(t, fixture, evtype)
}
