package event

import (
	"testing"

	"github.com/xytis/gami"
)

func TestRenameEvent(t *testing.T) {
	fixture := map[string]string{
		"Channel":  "Channel",
		"Newname":  "NewName",
		"Uniqueid": "UniqueID",
	}

	ev := gami.AMIEvent{
		ID:        "Rename",
		Privilege: []string{"all"},
		Params:    fixture,
	}

	evtype := New(&ev)
	if _, ok := evtype.(Rename); !ok {
		t.Log("Rename type assertion")
		t.Fail()
	}

	testEvent(t, fixture, evtype)
}
