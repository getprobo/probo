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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KeyFileName stores the device API key on disk.
const KeyFileName = "agent.key"

// ErrKeyNotFound is returned when no usable key is on disk (missing,
// empty, or whitespace-only file).
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
	if err := replaceRegularFile(path, []byte(strings.TrimSpace(key)+"\n"), 0o600); err != nil {
		return fmt.Errorf("cannot replace key: %w", err)
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

	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", ErrKeyNotFound
	}

	return key, nil
}

// DeleteAPIKey removes the API key file.
func DeleteAPIKey(dir string) error {
	if err := os.Remove(KeyPath(dir)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot delete agent key: %w", err)
	}

	return nil
}
