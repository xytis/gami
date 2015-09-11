GAMI
====

GO - Asterisk AMI Interface

communicate with the Asterisk AMI, Actions and Events.

Example connecting to Asterisk and Send Action get Events.

```go
package main
import (
	"log"
	"github.com/xytis/gami"
	"github.com/xytis/gami/event"
)

func main() {
	ami, err := gami.Connect("127.0.0.1:5038", "admin", "root")
	if err != nil {
		log.Fatal(err)
	}

	//install manager
	go func() {
		for {
			select {
			case ev, ok := <-ami.Events:
				if !ok {
					return
				}
				log.Println("received an event", ev, ok)
				log.Println("event type:", event.New(ev))
			case err, ok := <-ami.Errors:
				log.Println("received an error", err, ok)
			case fatal, ok := <-ami.Fatal:
				log.Println("ami stack died with", fatal, ok)
				log.Println("should try to recover...")
				return
			}
		}
	}()

	if rs, err = ami.Action("Ping", nil); err != nil {
		log.Fatal(rs)
	}

	//async actions
	rsPing, rsErr := ami.AsyncAction("Ping", gami.Params{"ActionID": "pingo"})
	if rsErr != nil {
		log.Fatal(rsErr)
	}

	if rs, err = ami.Action("Events", ami.Params{"EventMask":"on"}); err != nil {
		log.Fatal(err)
	}

	log.Println("ping:", <-rsPing)

	ami.Close()
}
```

CURRENT EVENT TYPES
====

The events use documentation and struct from *PAMI*.

use **xytis/gami/event.New()** for get this struct from raw event

EVENT ID           | TYPE TEST
------------------ | ----------
*Newchannel*       | YES
*Newexten*         | YES
*Newstate*         | YES
*Dial*             | YES
*ExtensionStatus*  | YES
*Hangup*           | YES
*PeerStatus*       | YES
*PeerEntry*        | YES
*VarSet*           | YES
*AgentLogin*       | YES
*Agents*           | YES
*AgentLogoff*      | YES
*AgentConnect*     | YES
*RTPReceiverStats* | YES
*RTPSenderStats*   | YES
*Bridge*           | YES
