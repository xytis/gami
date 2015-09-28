// Package event for AMI
package event

// Raised when a channel joins a Queue.
type Join struct {
	Privilege         []string
	Queue             string `AMI:"Queue"`
	Position          string `AMI:"Position"`
	Count             string `AMI:"Count"`
	Channel           string `AMI:"Channel"`
	CallerIDNum       string `AMI:"CallerIDNum"`
	CallerIDName      string `AMI:"CallerIDName"`
	ConnectedLineNum  string `AMI:"ConnectedLineNum"`
	ConnectedLineName string `AMI:"ConnectedLineName"`
	UniqueID          string `AMI:"Uniqueid"`
}

func init() {
	eventTrap["Join"] = Join{}
}
