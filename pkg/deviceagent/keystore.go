// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package deviceagent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KeyFileName stores the device API key on disk.
const KeyFileName = "agent.key"

// ErrKeyNotFound is returned when no key file exists.
var ErrKeyNotFound = errors.New("agent key not found")

// KeyPath returns the absolute path of the device API key file.
func KeyPath(dir string) string {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	return filepath.Join(dir, KeyFileName)
}

// SaveAPIKey writes the API key to disk with mode 0600.
func SaveAPIKey(dir, key string) error {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("cannot create keystore dir: %w", err)
	}

	path := KeyPath(dir)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.TrimSpace(key)+"\n"), 0o600); err != nil {
		return fmt.Errorf("cannot write key: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("cannot atomically replace key: %w", err)
	}

	return nil
}

// LoadAPIKey reads the API key from disk.
func LoadAPIKey(dir string) (string, error) {
	data, err := os.ReadFile(KeyPath(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrKeyNotFound
		}

		return "", fmt.Errorf("cannot read agent key: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// DeleteAPIKey removes the API key file.
func DeleteAPIKey(dir string) error {
	if err := os.Remove(KeyPath(dir)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot delete agent key: %w", err)
	}

	return nil
}
