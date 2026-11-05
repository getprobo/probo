// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
	// EnrollmentsFileName stores one API key per organization enrollment.
	EnrollmentsFileName = "enrollments.json"
)

type (
	// Enrollment is one organization registration of this machine.
	Enrollment struct {
		ServerURL string `json:"server_url"`
		DeviceID  string `json:"device_id,omitempty"`
		APIKey    string `json:"api_key"`
	}
)

// EnrollmentsPath returns the absolute path of the enrollment secret file.
func EnrollmentsPath(dir string) string {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	return filepath.Join(dir, EnrollmentsFileName)
}

// LoadEnrollments reads enrollments.json. When that file is absent and a
// legacy agent.key plus config server URL exist, they are migrated into
// enrollments.json and the legacy key is removed.
func LoadEnrollments(dir string) ([]Enrollment, error) {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	items, found, err := readEnrollmentFile(dir)
	if err != nil {
		return nil, err
	}

	if found {
		if err := DeleteAPIKey(dir); err != nil {
			return nil, fmt.Errorf("cannot delete legacy agent key: %w", err)
		}

		return items, nil
	}

	return migrateLegacyEnrollment(dir)
}

// SaveEnrollments writes the enrollment secret file, or removes it when
// the list is empty.
func SaveEnrollments(dir string, items []Enrollment) error {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	if err := ensureSecureAgentDir(dir); err != nil {
		return err
	}

	path := EnrollmentsPath(dir)
	if len(items) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("cannot delete enrollments: %w", err)
		}

		return nil
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode enrollments: %w", err)
	}

	data = append(data, '\n')
	if err := replaceRegularFile(path, data, 0o600); err != nil {
		return fmt.Errorf("cannot replace enrollments: %w", err)
	}

	return nil
}

func syncPublicEnrollmentState(dir string, items []Enrollment) error {
	runDir := EnrollmentRunDir(dir)
	if len(items) == 0 {
		return ClearEnrollmentMarker(runDir)
	}

	return MarkEnrolled(runDir)
}

func readEnrollmentFile(dir string) ([]Enrollment, bool, error) {
	data, err := os.ReadFile(EnrollmentsPath(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("cannot read enrollments: %w", err)
	}

	var items []Enrollment
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, false, fmt.Errorf("cannot decode enrollments: %w", err)
	}

	filtered := make([]Enrollment, 0, len(items))
	for _, item := range items {
		if item.ServerURL == "" || item.APIKey == "" {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered, true, nil
}

func migrateLegacyEnrollment(dir string) ([]Enrollment, error) {
	apiKey, err := LoadAPIKey(dir)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load legacy agent key: %w", err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("legacy agent key has no server URL")
		}

		return nil, fmt.Errorf("cannot load config for enrollment migration: %w", err)
	}

	if cfg.ServerURL == "" {
		return nil, errors.New("legacy agent key has no server URL")
	}

	normalized, err := NormalizeServerURL(cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid legacy server URL: %w", err)
	}

	item := Enrollment{
		ServerURL: normalized,
		DeviceID:  cfg.DeviceID,
		APIKey:    apiKey,
	}
	if err := SaveEnrollments(dir, []Enrollment{item}); err != nil {
		return nil, err
	}

	cfg.ServerURL = ""

	cfg.DeviceID = ""
	if err := SaveConfig(dir, cfg); err != nil {
		return nil, fmt.Errorf("cannot clear legacy enrollment config: %w", err)
	}

	if err := DeleteAPIKey(dir); err != nil {
		return nil, fmt.Errorf("cannot delete legacy agent key: %w", err)
	}

	return []Enrollment{item}, nil
}

func loadOrCreateConfig(dir string) (*Config, error) {
	cfg, err := LoadConfig(dir)
	if err == nil {
		return cfg, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("cannot load config: %w", err)
	}

	cfg = &Config{}
	if err := SaveConfig(dir, cfg); err != nil {
		return nil, fmt.Errorf("cannot create config: %w", err)
	}

	return LoadConfig(dir)
}

// serverInterval keeps current when the server omits a cadence.
// Otherwise it returns the reported cadence, raised to floor.
func serverInterval(current, reported, floor time.Duration) time.Duration {
	if reported <= 0 {
		return current
	}

	return max(reported, floor)
}
