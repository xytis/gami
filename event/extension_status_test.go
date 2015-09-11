package event

import (
	"testing"

	"github.com/xytis/gami"
)

func TestExtensionStatus(t *testing.T) {
	fixture := map[string]string{
		"Exten":   "Extension",
		"Context": "Context",
		"Hint":    "Hint",
		"Status":  "Status",
	}

	ev := gami.AMIEvent{
		ID:        "ExtensionStatus",
		Privilege: []string{"all"},
		Params:    fixture,
	}

	evtype := New(&ev)

	if _, ok := evtype.(ExtensionStatus); !ok {
		t.Fatal("ExtensionStatus type assertion")
	}

	testEvent(t, fixture, evtype)
}
