// Package event for AMI
package event

// Raised when the name of a channel is changed.
type Rename struct {
	Privilege []string
	Channel   string `AMI:"Channel"`
	NewName   string `AMI:"Newname"`
	UniqueID  string `AMI:"Uniqueid"`
}

func init() {
	eventTrap["Rename"] = Rename{}
}
