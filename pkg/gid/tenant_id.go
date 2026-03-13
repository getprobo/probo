package gid

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// TenantID represents a tenant identifier (64 bits/8 bytes total)
type TenantID [8]byte

var (
	// NilTenant represents an empty tenant ID
	NilTenant = TenantID{}

	// Global generation singleton
	defaultTenantGenerator = newTenantGenerator()
)

// TenantGenerator handles creation of unique tenant IDs
type tenantGenerator struct {
	// Process-specific values
	machineID [3]byte // 24 bits for machine identifier
	counter   uint32  // Counter for the sequence
}

// NewTenantID generates a new globally unique tenant ID
func NewTenantID() TenantID {
	return defaultTenantGenerator.NewTenantID()
}

// newTenantGenerator creates a new generator with machine-specific components
func newTenantGenerator() *tenantGenerator {
	g := &tenantGenerator{
		counter: 0,
	}

	// Generate machine ID component
	if _, err := rand.Read(g.machineID[:]); err != nil {
		// Fallback if random source fails
		hostname, _ := os.Hostname()
		copy(g.machineID[:], []byte(hostname))

		// Pad with timestamp bits if hostname is short
		if len(hostname) < len(g.machineID) {
			ts := time.Now().UnixNano()
			binary.BigEndian.PutUint16(g.machineID[len(hostname):], uint16(ts))
		}
	}

	return g
}

// NewTenantID generates a new 64-bit tenant ID with the structure:
// - 24 bits: Machine ID (random, unique per machine)
// - 24 bits: Timestamp (truncated Unix time in seconds)
// - 16 bits: Counter (increments per ID)
func (g *tenantGenerator) NewTenantID() TenantID {
	// Create new ID
	var id TenantID

	// 1. Copy machine ID (first 3 bytes)
	copy(id[0:3], g.machineID[:])

	// 2. Add timestamp (next 3 bytes) - Unix time in seconds (truncated)
	// Using seconds instead of milliseconds gives us until year 2286
	now := uint32(time.Now().Unix())
	id[3] = byte(now >> 16)
	id[4] = byte(now >> 8)
	id[5] = byte(now)

	// 3. Increment counter atomically (last 2 bytes)
	count := atomic.AddUint32(&g.counter, 1) & 0xFFFF
	binary.BigEndian.PutUint16(id[6:8], uint16(count))

	return id
}

// Value implements the database/sql/driver.Valuer interface
func (id TenantID) Value() (driver.Value, error) {
	return id.String(), nil
}

// Scan implements the database/sql.Scanner interface
func (id *TenantID) Scan(value any) error {
	switch v := value.(type) {
	case string:
		decoded, err := base64.RawURLEncoding.DecodeString(v)
		if err != nil {
			return err
		}

		if len(decoded) != len(*id) {
			return fmt.Errorf("invalid tenant ID length: got %d, want %d", len(decoded), len(*id))
		}

		copy((*id)[:], decoded)
		return nil
	default:
		return fmt.Errorf("invalid type for TenantID: expected string, got %T", value)
	}
}

// String returns the base64 representation of the TenantID
func (id TenantID) String() string {
	return base64.RawURLEncoding.EncodeToString(id[:])
}

// MarshalText returns the base64 representation for JSON encoding
func (id TenantID) MarshalText() ([]byte, error) {
	encoded := base64.RawURLEncoding.EncodeToString(id[:])
	return []byte(encoded), nil
}

// UnmarshalText parses the base64 representation for JSON decoding
func (id *TenantID) UnmarshalText(text []byte) error {
	decoded, err := base64.RawURLEncoding.DecodeString(string(text))
	if err != nil {
		return err
	}

	if len(decoded) != len(*id) {
		return fmt.Errorf("invalid tenant ID length: got %d, want %d", len(decoded), len(*id))
	}

	copy((*id)[:], decoded)
	return nil
}

// IsValid returns true if the tenant ID is not nil
func (id TenantID) IsValid() bool {
	return id != NilTenant
}

// Timestamp extracts the timestamp from the TenantID
func (id TenantID) Timestamp() time.Time {
	seconds := uint32(id[3])<<16 | uint32(id[4])<<8 | uint32(id[5])
	return time.Unix(int64(seconds), 0)
}
