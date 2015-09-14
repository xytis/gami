// Package event for AMI
package event

// Raised when Asterisk is shutdown or restarted.
type Shutdown struct {
	Privilege []string
	Shutdown  string `AMI:"Shutdown"`
	Restart   string `AMI:"Restart"`
}

func init() {
	eventTrap["Shutdown"] = Shutdown{}
}
