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
	pendingPosturesFileName  = "pending-postures.json"
	maxPendingPostureBatches = 96
)

type pendingPostureBatch struct {
	QueuedAt time.Time              `json:"queued_at"`
	Results  []PostureResultPayload `json:"results"`
}

func pendingPosturesPath(dir string) string {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	return filepath.Join(dir, pendingPosturesFileName)
}

func loadPendingPostureBatches(dir string) ([]pendingPostureBatch, error) {
	data, err := os.ReadFile(pendingPosturesPath(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot read pending postures: %w", err)
	}

	var batches []pendingPostureBatch
	if err := json.Unmarshal(data, &batches); err != nil {
		return nil, fmt.Errorf("cannot decode pending postures: %w", err)
	}

	filtered := make([]pendingPostureBatch, 0, len(batches))
	for _, batch := range batches {
		if len(batch.Results) == 0 {
			continue
		}

		filtered = append(filtered, batch)
	}

	return filtered, nil
}

func savePendingPostureBatches(dir string, batches []pendingPostureBatch) error {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	path := pendingPosturesPath(dir)
	if len(batches) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("cannot delete pending postures: %w", err)
		}

		return nil
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("cannot create pending posture dir: %w", err)
	}

	data, err := json.MarshalIndent(batches, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode pending postures: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("cannot write pending postures: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("cannot atomically replace pending postures: %w", err)
	}

	return nil
}

func enqueuePendingPostureBatch(
	dir string,
	results []PostureResultPayload,
	queuedAt time.Time,
) (int, error) {
	if len(results) == 0 {
		return 0, nil
	}

	batches, err := loadPendingPostureBatches(dir)
	if err != nil {
		return 0, err
	}

	clonedResults := make([]PostureResultPayload, len(results))
	copy(clonedResults, results)

	batches = append(
		batches,
		pendingPostureBatch{
			QueuedAt: queuedAt.UTC(),
			Results:  clonedResults,
		},
	)

	dropped := 0
	if len(batches) > maxPendingPostureBatches {
		dropped = len(batches) - maxPendingPostureBatches
		batches = batches[dropped:]
	}

	if err := savePendingPostureBatches(dir, batches); err != nil {
		return 0, err
	}

	return dropped, nil
}

func clearPendingPostureBatches(dir string) error {
	return savePendingPostureBatches(dir, nil)
}
