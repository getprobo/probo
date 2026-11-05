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
	"strings"
	"time"
)

const (
	pendingPosturesFileName  = "pending-postures.json"
	maxPendingPostureBatches = 96
)

type pendingPostureBatch struct {
	QueuedAt     time.Time              `json:"queued_at"`
	AgentVersion string                 `json:"agent_version,omitempty"`
	Results      []PostureResultPayload `json:"results"`
}

func pendingPosturesPath(dir string) string {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	return filepath.Join(dir, pendingPosturesFileName)
}

type pendingPostureFile struct {
	ByDevice map[string][]pendingPostureBatch `json:"by_device"`
}

func loadPendingPostureBatches(dir, deviceID string) ([]pendingPostureBatch, error) {
	queues, err := loadPostureQueues(dir)
	if err != nil {
		return nil, err
	}

	return filterPostureBatches(queues[deviceID]), nil
}

func loadPostureQueues(dir string) (map[string][]pendingPostureBatch, error) {
	path := pendingPosturesPath(dir)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string][]pendingPostureBatch{}, nil
		}

		return nil, fmt.Errorf("cannot read pending postures: %w", err)
	}

	// A leading '[' is the v0.8.0 queue, a JSON array with no device id.
	// Those batches are dropped; the next posture run starts clean.
	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) > 0 && trimmed[0] == '[' {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("cannot delete pending postures: %w", err)
		}

		return map[string][]pendingPostureBatch{}, nil
	}

	var file pendingPostureFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("cannot decode pending postures: %w", err)
	}

	if file.ByDevice == nil {
		file.ByDevice = map[string][]pendingPostureBatch{}
	}

	for deviceID, batches := range file.ByDevice {
		file.ByDevice[deviceID] = filterPostureBatches(batches)
		if len(file.ByDevice[deviceID]) == 0 {
			delete(file.ByDevice, deviceID)
		}
	}

	return file.ByDevice, nil
}

func savePendingPostureBatches(dir, deviceID string, batches []pendingPostureBatch) error {
	queues, err := loadPostureQueues(dir)
	if err != nil {
		return err
	}

	batches = filterPostureBatches(batches)
	if len(batches) == 0 {
		delete(queues, deviceID)
	} else {
		queues[deviceID] = batches
	}

	return savePostureQueues(dir, queues)
}

func savePostureQueues(dir string, queues map[string][]pendingPostureBatch) error {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	path := pendingPosturesPath(dir)
	if len(queues) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("cannot delete pending postures: %w", err)
		}

		return nil
	}

	if err := ensureSecureAgentDir(dir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(pendingPostureFile{ByDevice: queues}, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode pending postures: %w", err)
	}

	if err := replaceRegularFile(path, data, 0o600); err != nil {
		return fmt.Errorf("cannot replace pending postures: %w", err)
	}

	return nil
}

func enqueuePendingPostureBatch(
	dir string,
	deviceID string,
	agentVersion string,
	results []PostureResultPayload,
	queuedAt time.Time,
) (int, error) {
	if len(results) == 0 || deviceID == "" {
		return 0, nil
	}

	batches, err := loadPendingPostureBatches(dir, deviceID)
	if err != nil {
		return 0, err
	}

	clonedResults := make([]PostureResultPayload, len(results))
	copy(clonedResults, results)

	batches = append(
		batches,
		pendingPostureBatch{
			QueuedAt:     queuedAt.UTC(),
			AgentVersion: agentVersion,
			Results:      clonedResults,
		},
	)

	dropped := 0
	if len(batches) > maxPendingPostureBatches {
		dropped = len(batches) - maxPendingPostureBatches
		batches = batches[dropped:]
	}

	if err := savePendingPostureBatches(dir, deviceID, batches); err != nil {
		return 0, err
	}

	return dropped, nil
}

func clearPendingPostureBatches(dir, deviceID string) error {
	return savePendingPostureBatches(dir, deviceID, nil)
}

func filterPostureBatches(batches []pendingPostureBatch) []pendingPostureBatch {
	filtered := make([]pendingPostureBatch, 0, len(batches))
	for _, batch := range batches {
		if len(batch.Results) == 0 {
			continue
		}

		filtered = append(filtered, batch)
	}

	return filtered
}
