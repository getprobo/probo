// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// Package deviceagent implements the probo host agent.
package deviceagent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// ConfigFileName stores persisted agent config.
	ConfigFileName = "config.json"

	// DefaultHeartbeatInterval is the default heartbeat cadence.
	DefaultHeartbeatInterval = 5 * time.Minute
	// MinHeartbeatInterval is the minimum heartbeat cadence.
	MinHeartbeatInterval = 1 * time.Minute

	// DefaultPostureInterval is the default posture cadence.
	DefaultPostureInterval = 1 * time.Hour
	// MinPostureInterval is the minimum posture cadence.
	MinPostureInterval = 15 * time.Minute

	// DefaultUpdateInterval is the default cadence at which the
	// agent checks for new releases.
	DefaultUpdateInterval = 4 * time.Hour
	// MinUpdateInterval is the floor used when a smaller value is
	// configured. Updates are network and disk heavy, so we cap
	// frequency to once per hour.
	MinUpdateInterval = 1 * time.Hour
)

type (
	// Config is the persisted agent configuration.
	Config struct {
		ServerURL         string        `json:"server_url"`
		DeviceID          string        `json:"device_id,omitempty"`
		HeartbeatInterval time.Duration `json:"heartbeat_interval,omitempty"`
		PostureInterval   time.Duration `json:"posture_interval,omitempty"`
		UpdateInterval    time.Duration `json:"update_interval,omitempty"`
		UpdatesDisabled   bool          `json:"updates_disabled,omitempty"`
	}
)

// ConfigPath returns the absolute path to the agent's config file.
func ConfigPath(dir string) string {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	return filepath.Join(dir, ConfigFileName)
}

// LoadConfig reads config from disk.
func LoadConfig(dir string) (*Config, error) {
	data, err := os.ReadFile(ConfigPath(dir))
	if err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("cannot decode config: %w", err)
	}

	cfg.applyDefaults()

	return cfg, nil
}

// SaveConfig writes config to disk with mode 0600.
func SaveConfig(dir string, cfg *Config) error {
	if cfg == nil {
		return errors.New("nil config")
	}

	if dir == "" {
		dir = DefaultConfigDir()
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	cfg.applyDefaults()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode config: %w", err)
	}

	path := ConfigPath(dir)

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("cannot atomically replace config: %w", err)
	}

	return nil
}

func (c *Config) applyDefaults() {
	c.HeartbeatInterval = normalizeHeartbeatInterval(c.HeartbeatInterval)
	c.PostureInterval = normalizePostureInterval(c.PostureInterval)
	c.UpdateInterval = normalizeUpdateInterval(c.UpdateInterval)
}

func normalizeHeartbeatInterval(v time.Duration) time.Duration {
	return normalizeInterval(v, DefaultHeartbeatInterval, MinHeartbeatInterval)
}

func normalizePostureInterval(v time.Duration) time.Duration {
	return normalizeInterval(v, DefaultPostureInterval, MinPostureInterval)
}

func normalizeUpdateInterval(v time.Duration) time.Duration {
	return normalizeInterval(v, DefaultUpdateInterval, MinUpdateInterval)
}

func normalizeInterval(v, fallback, floor time.Duration) time.Duration {
	if v <= 0 {
		v = fallback
	}

	if v < floor {
		return floor
	}

	return v
}
