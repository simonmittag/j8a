package server

//ResourceMapping describes upstream servers
type ResourceMapping struct {
	Name   string
	Labels []string
	URL    URL
}
