// Package event for AMI
package event

// A user defined event raised from the dialplan.
type UserEvent struct {
	Privilege []string
	UserEvent string `AMI:"UserEvent"`
	UniqueID  string `AMI:"Uniqueid"`
}

func init() {
	eventTrap["UserEvent"] = UserEvent{}
}
