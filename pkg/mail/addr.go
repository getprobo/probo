package mail

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
)

type Addr string

var Nil Addr

func (a Addr) String() string {
	return string(a)
}

func (a *Addr) Username() string {
	if a == nil || *a == Nil {
		return ""
	}

	parts := strings.Split(a.String(), "@")
	if len(parts) != 2 {
		return ""
	}

	return parts[0]
}

func (a *Addr) Domain() string {
	if a == nil || *a == Nil {
		return ""
	}

	parts := strings.Split(a.String(), "@")
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}

func ParseAddr(s string) (Addr, error) {
	netMailAddr, err := mail.ParseAddress(s)
	if err != nil {
		return Nil, fmt.Errorf("cannot parse address %s: %w", s, err)
	}

	parts := strings.Split(netMailAddr.Address, "@")
	if len(parts) != 2 {
		return Nil, fmt.Errorf("invalid email address format")
	}

	return Addr(netMailAddr.Address), nil
}

func (a Addr) Value() (driver.Value, error) {
	return a.String(), nil
}

func (a *Addr) Scan(value any) error {
	if value == nil {
		*a = Nil
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("invalid type %T for mail.Addr", value)
	}

	parsed, err := ParseAddr(str)
	if err != nil {
		return err
	}

	*a = parsed

	return nil
}

func (a *Addr) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal email string from JSON")
	}

	parsed, err := ParseAddr(s)
	if err != nil {
		return fmt.Errorf("cannot parse email address")
	}

	*a = parsed

	return nil
}

func (a Addr) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

type Addrs []Addr

func (a Addrs) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	if len(a) == 0 {
		return "{}", nil
	}
	strs := make([]string, len(a))
	for i, addr := range a {
		strs[i] = addr.String()
	}
	return "{" + strings.Join(strs, ",") + "}", nil
}

func (a *Addrs) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}

	var strs []string
	switch v := value.(type) {
	case []string:
		strs = v
	case []byte:
		s := strings.Trim(string(v), "{}")
		if s == "" {
			strs = []string{}
		} else {
			strs = strings.Split(s, ",")
		}
	case string:
		s := strings.Trim(v, "{}")
		if s == "" {
			strs = []string{}
		} else {
			strs = strings.Split(s, ",")
		}
	case []any:
		strs = make([]string, len(v))
		for i, elem := range v {
			if elem == nil {
				strs[i] = ""
				continue
			}
			str, ok := elem.(string)
			if !ok {
				return fmt.Errorf("array element is not a string: %T", elem)
			}
			strs[i] = str
		}
	default:
		return fmt.Errorf("cannot scan %T into Addrs", value)
	}

	*a = make([]Addr, len(strs))
	for i, str := range strs {
		if str == "" {
			(*a)[i] = Nil
			continue
		}
		parsed, err := ParseAddr(str)
		if err != nil {
			return fmt.Errorf("invalid email at index %d: %w", i, err)
		}
		(*a)[i] = parsed
	}
	return nil
}
