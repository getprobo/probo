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

	"go.probo.inc/probo/pkg/deviceagent/checks"
)

const checkMemoryFileName = "check-memory.json"

// checkMemory is the last PASS or FAIL per check key.
type checkMemory map[string]checks.Status

func checkMemoryPath(dir string) string {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	return filepath.Join(dir, checkMemoryFileName)
}

func loadCheckMemory(dir string) (checkMemory, error) {
	data, err := os.ReadFile(checkMemoryPath(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return checkMemory{}, nil
		}

		return nil, fmt.Errorf("cannot read check memory: %w", err)
	}

	var memory checkMemory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, fmt.Errorf("cannot decode check memory: %w", err)
	}

	if memory == nil {
		return checkMemory{}, nil
	}

	return memory, nil
}

func saveCheckMemory(dir string, memory checkMemory) error {
	path := checkMemoryPath(dir)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("cannot create check memory dir: %w", err)
	}

	data, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode check memory: %w", err)
	}

	if err := replaceRegularFile(path, data, 0o600); err != nil {
		return fmt.Errorf("cannot replace check memory: %w", err)
	}

	return nil
}

func applyCheckMemory(results []checks.Result, memory checkMemory) bool {
	dirty := false

	for i := range results {
		result := &results[i]

		switch result.Status {
		case checks.StatusPass, checks.StatusFail:
			if memory[result.CheckKey] != result.Status {
				memory[result.CheckKey] = result.Status
				dirty = true
			}
		case checks.StatusUnknown:
			if !result.Rememberable {
				continue
			}

			last, ok := memory[result.CheckKey]
			if !ok || (last != checks.StatusPass && last != checks.StatusFail) {
				continue
			}

			result.Status = last
			if result.Evidence == nil {
				result.Evidence = map[string]any{}
			}

			result.Evidence["remembered"] = true
			if result.CheckKey == checks.KeyScreenLock {
				result.Evidence["screen_lock_enforced"] = last == checks.StatusPass
			}
		}
	}

	return dirty
}
