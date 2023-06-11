package j8a

// URL describes host mapping
type URL struct {
	Scheme string
	Host   string
	Port   string
}

// String representation of our URL struct
func (u URL) String() string {
	return u.Scheme + "://" + u.Host + ":" + u.Port
}
