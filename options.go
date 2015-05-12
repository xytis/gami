package gami

// Options
func UseTLS(c *AMIClient) {
	c.useTLS = true
}

func UnsecureTLS(c *AMIClient) {
	c.unsecureTLS = true
}
