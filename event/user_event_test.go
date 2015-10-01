package event

import (
	"testing"

	"github.com/xytis/gami"
)

func TestUserEventEvent(t *testing.T) {
	fixture := map[string]string{
		"Userevent": "UserEvent",
		"Uniqueid":  "UniqueID",
	}

	ev := gami.AMIEvent{
		ID:        "UserEvent",
		Privilege: []string{"all"},
		Params:    fixture,
	}

	evtype := New(&ev)
	if _, ok := evtype.(UserEvent); !ok {
		t.Log("UserEvent type assertion")
		t.Fail()
	}

	testEvent(t, fixture, evtype)
}
