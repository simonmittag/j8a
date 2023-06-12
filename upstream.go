package j8a

import (
	"encoding/json"
	"fmt"
)

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

func (u *URL) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case map[string]interface{}:
		if v["scheme"] != nil {
			u.Scheme = fmt.Sprintf("%v", v["scheme"])
		}
		if v["host"] != nil {
			u.Host = fmt.Sprintf("%v", v["host"])
		}
		if v["port"] != nil {
			u.Port = fmt.Sprintf("%v", v["port"])
		}
	default:
		return fmt.Errorf("unexpected JSON value type: %T", value)
	}

	return nil
}
