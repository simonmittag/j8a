package server

//ResourceMapping describes upstream servers
type ResourceMapping struct {
	Alias    string
	Labels   []string
	Upstream Upstream
}
