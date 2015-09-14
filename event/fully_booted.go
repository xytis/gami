// Package event for AMI
package event

// Raised when all Asterisk initialization procedures have finished.
type FullyBooted struct {
	Privilege []string
	Status    string `AMI:"Status"`
}

func init() {
	eventTrap["FullyBooted"] = FullyBooted{}
}
