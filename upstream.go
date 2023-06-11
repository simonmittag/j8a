package j8a

import (
	"strconv"
)

// URL describes host mapping
type URL struct {
	Scheme string
	Host   string
	Port   int
}

// String representation of our URL struct
func (u URL) String() string {
	return u.Scheme + "://" + u.Host + ":" + strconv.Itoa(int(u.Port))
}
