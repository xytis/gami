// Package event for AMI
package event

// Raised when a channel leaves a Queue.
type Leave struct {
	Privilege []string
	Queue     string `AMI:"Queue"`
	Count     string `AMI:"Count"`
	Position  string `AMI:"Position"`
	Channel   string `AMI:"Channel"`
	UniqueID  string `AMI:"Uniqueid"`
}

func init() {
	eventTrap["Leave"] = Leave{}
}
