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

func (a *Addr) Domain() string {
	return strings.Split(a.String(), "@")[1]
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
	switch v := value.(type) {
	case string:
		parsed, err := ParseAddr(v)
		if err != nil {
			return err
		}

		*a = parsed
	default:
		return fmt.Errorf("invalid type for mail.Addr: expected string, got %T", value)
	}
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
