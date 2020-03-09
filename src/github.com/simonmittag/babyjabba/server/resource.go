package server

//Resource describes upstream servers
type Resource struct {
	Alias    string
	Labels   []string
	Upstream Upstream
}
