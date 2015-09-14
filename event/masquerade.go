// Package event for AMI
package event

// Raised when a masquerade occurs between two channels, wherein the Clone channel's internal information replaces the Original channel's information.
type Masquerade struct {
	Privilege     []string
	Clone         string `AMI:"Clone"`
	CloneState    string `AMI:"CloneState"`
	Original      string `AMI:"Original"`
	OriginalState string `AMI:"OriginalState"`
}

func init() {
	eventTrap["Masquerade"] = Masquerade{}
}
