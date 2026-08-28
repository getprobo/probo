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

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/benchmarkmetrics"
)

type (
	result struct {
		Workload            string                   `json:"workload"`
		Size                int                      `json:"size"`
		Iterations          int                      `json:"iterations"`
		TotalNS             int64                    `json:"totalNs"`
		NSPerOp             int64                    `json:"nsPerOp"`
		AllocsPerOp         uint64                   `json:"allocsPerOp"`
		BytesAllocatedPerOp uint64                   `json:"bytesAllocatedPerOp"`
		Checksum            string                   `json:"checksum"`
		OutputBytes         int                      `json:"outputBytes,omitempty"`
		OutputHash          string                   `json:"outputHash,omitempty"`
		Metrics             benchmarkmetrics.Metrics `json:"metrics"`
	}

	benchmarkWorkload struct {
		run      func() error
		validate func() (string, error)
		output   func() []byte
		cleanup  func()
	}
)

var (
	benchmarkActor = automerge.ActorID{
		0, 1, 2, 3, 4, 5, 6, 7,
		8, 9, 10, 11, 12, 13, 14, 15,
	}
	benchmarkPeerActor = automerge.ActorID{
		15, 14, 13, 12, 11, 10, 9, 8,
		7, 6, 5, 4, 3, 2, 1, 0,
	}
)

func main() {
	workload := flag.String("workload", "", "benchmark workload")
	size := flag.Int("size", 0, "workload size")
	iterations := flag.Int("iterations", 1, "timed iterations")
	warmups := flag.Int("warmups", 3, "warmup iterations")
	fixture := flag.String("fixture", "", "shared fixture path")

	flag.Parse()

	if *workload == "" || *size < 0 || *iterations <= 0 || *warmups < 0 {
		fmt.Fprintln(os.Stderr, "invalid benchmark arguments")
		os.Exit(2)
	}

	if *workload == "fixture" {
		if *fixture == "" {
			fmt.Fprintln(os.Stderr, "fixture path is required")
			os.Exit(2)
		}

		document, err := fixtureDocument(*size)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		data, err := document.Save()
		_ = document.Close()

		if err == nil {
			err = os.WriteFile(*fixture, data, 0o600)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return
	}

	runner, err := workloadRunner(*workload, *size, *fixture)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer runner.cleanup()

	for range *warmups {
		if err := runner.run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	benchmarkmetrics.Reset()
	startedAt := time.Now()

	for range *iterations {
		if err := runner.run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	total := time.Since(startedAt)
	metrics := benchmarkmetrics.Read()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	checksum, err := runner.validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	output := runner.output()

	if err := json.NewEncoder(os.Stdout).Encode(
		result{
			Workload:            *workload,
			Size:                *size,
			Iterations:          *iterations,
			TotalNS:             total.Nanoseconds(),
			NSPerOp:             total.Nanoseconds() / int64(*iterations),
			AllocsPerOp:         (after.Mallocs - before.Mallocs) / uint64(*iterations),
			BytesAllocatedPerOp: (after.TotalAlloc - before.TotalAlloc) / uint64(*iterations),
			Checksum:            checksum,
			OutputBytes:         len(output),
			OutputHash:          optionalChecksum(output),
			Metrics:             metrics,
		},
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func workloadRunner(
	workload string,
	size int,
	fixture string,
) (benchmarkWorkload, error) {
	emptyOutput := func() []byte { return nil }

	switch workload {
	case "create":
		return benchmarkWorkload{
			run: func() error {
				document, err := automerge.New(benchmarkActor)
				if err != nil {
					return err
				}

				return document.Close()
			},
			validate: func() (string, error) {
				return checksum([]byte("empty")), nil
			},
			output:  emptyOutput,
			cleanup: func() {},
		}, nil
	case "map":
		return benchmarkWorkload{
			run: func() error {
				document, err := mapDocument(size)
				if err != nil {
					return err
				}

				return document.Close()
			},
			validate: func() (string, error) {
				document, err := mapDocument(size)
				if err != nil {
					return "", err
				}

				defer func() { _ = document.Close() }()

				return mapChecksum(document, size)
			},
			output:  emptyOutput,
			cleanup: func() {},
		}, nil
	case "map-update":
		document, err := mapDocument(size)
		if err != nil {
			return benchmarkWorkload{}, err
		}
		values, err := document.Root().Object("values")
		if err != nil {
			_ = document.Close()
			return benchmarkWorkload{}, err
		}

		return benchmarkWorkload{
			run: func() error {
				for index := range size {
					if err := values.PutScalar(
						strconv.Itoa(index),
						automerge.Scalar{
							Type: automerge.ScalarTypeInt,
							Int:  int64(size - index),
						},
					); err != nil {
						return err
					}
				}

				_, err := document.Commit("benchmark", time.Time{})

				return err
			},
			validate: func() (string, error) {
				return mapChecksum(document, size)
			},
			output: emptyOutput,
			cleanup: func() {
				_ = document.Close()
			},
		}, nil
	case "text":
		return benchmarkWorkload{
			run: func() error {
				document, err := typedDocument(size)
				if err != nil {
					return err
				}

				return document.Close()
			},
			validate: func() (string, error) {
				document, err := typedDocument(size)
				if err != nil {
					return "", err
				}

				defer func() { _ = document.Close() }()

				return textChecksum(document)
			},
			output:  emptyOutput,
			cleanup: func() {},
		}, nil
	case "text-edit":
		document, err := fixtureDocument(size)
		if err != nil {
			return benchmarkWorkload{}, err
		}
		text, err := document.Text("body")
		if err != nil {
			_ = document.Close()
			return benchmarkWorkload{}, err
		}

		return benchmarkWorkload{
			run: func() error {
				edits := min(size, 100)
				for index := range edits {
					position := uint32(index * size / edits)
					if err := text.Splice(position, 1, "z"); err != nil {
						return err
					}
				}

				_, err := document.Commit("benchmark", time.Time{})

				return err
			},
			validate: func() (string, error) {
				return textChecksum(document)
			},
			output: emptyOutput,
			cleanup: func() {
				_ = document.Close()
			},
		}, nil
	case "load":
		data, err := fixtureData(size, fixture)
		if err != nil {
			return benchmarkWorkload{}, err
		}

		return benchmarkWorkload{
			run: func() error {
				loaded, err := automerge.Load(
					data,
					benchmarkActor,
				)
				if err != nil {
					return err
				}

				return loaded.Close()
			},
			validate: func() (string, error) {
				loaded, err := automerge.Load(
					data,
					benchmarkActor,
				)
				if err != nil {
					return "", err
				}

				defer func() { _ = loaded.Close() }()

				return textChecksum(loaded)
			},
			output:  emptyOutput,
			cleanup: func() {},
		}, nil
	case "save":
		data, err := fixtureData(size, fixture)
		if err != nil {
			return benchmarkWorkload{}, err
		}

		document, err := automerge.Load(
			data,
			benchmarkActor,
		)
		if err != nil {
			return benchmarkWorkload{}, err
		}
		var latestSave []byte

		return benchmarkWorkload{
			run: func() error {
				latestSave, err = document.Save()

				return err
			},
			validate: func() (string, error) {
				data, err := document.Save()
				if err != nil {
					return "", err
				}

				loaded, err := automerge.Load(
					data,
					benchmarkActor,
				)
				if err != nil {
					return "", err
				}

				defer func() { _ = loaded.Close() }()

				return textChecksum(loaded)
			},
			output: func() []byte {
				return latestSave
			},
			cleanup: func() {
				_ = document.Close()
			},
		}, nil
	case "save-after-loaded-change", "save-after-change":
		data, err := fixtureData(size, fixture)
		if err != nil {
			return benchmarkWorkload{}, err
		}

		document, err := automerge.Load(
			data,
			benchmarkActor,
		)
		if err != nil {
			return benchmarkWorkload{}, err
		}

		text, err := document.Text("body")
		if err != nil {
			_ = document.Close()
			return benchmarkWorkload{}, err
		}

		position := uint32(size)
		var latestSave []byte

		return benchmarkWorkload{
			run: func() error {
				if err := text.Splice(position, 0, "x"); err != nil {
					return err
				}
				position++

				if _, err := document.Commit("benchmark", time.Time{}); err != nil {
					return err
				}

				latestSave, err = document.Save()

				return err
			},
			validate: func() (string, error) {
				return textChecksum(document)
			},
			output: func() []byte {
				return latestSave
			},
			cleanup: func() {
				_ = document.Close()
			},
		}, nil
	case "merge-loaded", "merge-reloaded", "concurrent-tail-reconcile":
		tailEdits := 1
		if workload == "concurrent-tail-reconcile" {
			tailEdits = min(size, 100)
		}
		leftData, rightData, err := mergeFixtureData(size, tailEdits)
		if err != nil {
			return benchmarkWorkload{}, err
		}

		merge := func() (*automerge.Document, error) {
			left, err := automerge.Load(leftData, benchmarkActor)
			if err != nil {
				return nil, err
			}

			right, err := automerge.Load(rightData, benchmarkPeerActor)
			if err != nil {
				_ = left.Close()
				return nil, err
			}
			defer func() { _ = right.Close() }()

			if _, err := left.Merge(right); err != nil {
				_ = left.Close()
				return nil, err
			}

			return left, nil
		}

		if workload == "merge-reloaded" {
			return benchmarkWorkload{
				run: func() error {
					document, err := merge()
					if err != nil {
						return err
					}

					return document.Close()
				},
				validate: func() (string, error) {
					document, err := merge()
					if err != nil {
						return "", err
					}
					defer func() { _ = document.Close() }()

					return textChecksum(document)
				},
				output:  emptyOutput,
				cleanup: func() {},
			}, nil
		}

		left, err := automerge.Load(leftData, benchmarkActor)
		if err != nil {
			return benchmarkWorkload{}, err
		}
		right, err := automerge.Load(rightData, benchmarkPeerActor)
		if err != nil {
			_ = left.Close()
			return benchmarkWorkload{}, err
		}

		return benchmarkWorkload{
			run: func() error {
				_, err := left.Merge(right)

				return err
			},
			validate: func() (string, error) {
				return textChecksum(left)
			},
			output: emptyOutput,
			cleanup: func() {
				_ = left.Close()
				_ = right.Close()
			},
		}, nil
	case "sync-initial":
		data, err := fixtureData(size, fixture)
		if err != nil {
			return benchmarkWorkload{}, err
		}
		var latest *automerge.Document
		var latestWire []byte

		return benchmarkWorkload{
			run: func() error {
				left, err := automerge.Load(data, benchmarkActor)
				if err != nil {
					return err
				}
				defer func() { _ = left.Close() }()

				right, err := automerge.New(benchmarkPeerActor)
				if err != nil {
					return err
				}

				leftState, err := left.NewSyncState()
				if err != nil {
					return err
				}
				defer func() { _ = leftState.Close() }()

				rightState, err := right.NewSyncState()
				if err != nil {
					return err
				}
				defer func() { _ = rightState.Close() }()

				latestWire, err = synchronizeWithWire(leftState, rightState)
				if err != nil {
					_ = right.Close()
					return err
				}
				if latest != nil {
					_ = latest.Close()
				}
				latest = right

				return nil
			},
			validate: func() (string, error) {
				if latest == nil {
					return "", fmt.Errorf("initial sync workload did not produce a document")
				}

				return textChecksum(latest)
			},
			output: func() []byte {
				return latestWire
			},
			cleanup: func() {
				if latest != nil {
					_ = latest.Close()
				}
			},
		}, nil
	case "sync-diverged":
		data, err := fixtureData(size, fixture)
		if err != nil {
			return benchmarkWorkload{}, err
		}
		left, err := automerge.Load(data, benchmarkActor)
		if err != nil {
			return benchmarkWorkload{}, err
		}
		right, err := automerge.Load(data, benchmarkPeerActor)
		if err != nil {
			_ = left.Close()
			return benchmarkWorkload{}, err
		}
		leftState, err := left.NewSyncState()
		if err != nil {
			_ = left.Close()
			_ = right.Close()
			return benchmarkWorkload{}, err
		}
		rightState, err := right.NewSyncState()
		if err != nil {
			_ = leftState.Close()
			_ = left.Close()
			_ = right.Close()
			return benchmarkWorkload{}, err
		}
		if err := synchronize(leftState, rightState); err != nil {
			return benchmarkWorkload{}, err
		}

		leftText, err := left.Text("body")
		if err != nil {
			return benchmarkWorkload{}, err
		}
		rightText, err := right.Text("body")
		if err != nil {
			return benchmarkWorkload{}, err
		}
		if err := leftText.Splice(uint32(size), 0, "L"); err != nil {
			return benchmarkWorkload{}, err
		}
		if _, err := left.Commit("benchmark", time.Time{}); err != nil {
			return benchmarkWorkload{}, err
		}
		if err := rightText.Splice(uint32(size), 0, "R"); err != nil {
			return benchmarkWorkload{}, err
		}
		if _, err := right.Commit("benchmark", time.Time{}); err != nil {
			return benchmarkWorkload{}, err
		}
		var latestWire []byte

		return benchmarkWorkload{
			run: func() error {
				var err error
				latestWire, err = synchronizeWithWire(leftState, rightState)

				return err
			},
			validate: func() (string, error) {
				leftChecksum, err := textChecksum(left)
				if err != nil {
					return "", err
				}
				rightChecksum, err := textChecksum(right)
				if err != nil {
					return "", err
				}
				if leftChecksum != rightChecksum {
					return "", fmt.Errorf("synchronized documents differ")
				}

				return leftChecksum, nil
			},
			output: func() []byte {
				return latestWire
			},
			cleanup: func() {
				_ = leftState.Close()
				_ = rightState.Close()
				_ = left.Close()
				_ = right.Close()
			},
		}, nil
	default:
		return benchmarkWorkload{}, fmt.Errorf("unknown workload %q", workload)
	}
}

func mapDocument(size int) (*automerge.Document, error) {
	document, err := automerge.New(benchmarkActor)
	if err != nil {
		return nil, err
	}

	values, err := document.Root().CreateObject(
		"values",
		automerge.ObjectTypeMap,
	)
	if err != nil {
		_ = document.Close()
		return nil, err
	}

	for index := range size {
		if err := values.PutScalar(
			strconv.Itoa(index),
			automerge.Scalar{
				Type: automerge.ScalarTypeInt,
				Int:  int64(index),
			},
		); err != nil {
			_ = document.Close()
			return nil, err
		}
	}

	if _, err := document.Commit("benchmark", time.Time{}); err != nil {
		_ = document.Close()
		return nil, err
	}

	return document, nil
}

func typedDocument(size int) (*automerge.Document, error) {
	document, err := automerge.New(benchmarkActor)
	if err != nil {
		return nil, err
	}

	text, err := document.CreateText("body")
	if err != nil {
		_ = document.Close()
		return nil, err
	}

	for index := range size {
		if err := text.Splice(uint32(index), 0, "x"); err != nil {
			_ = document.Close()
			return nil, err
		}
	}

	if _, err := document.Commit("benchmark", time.Time{}); err != nil {
		_ = document.Close()
		return nil, err
	}

	return document, nil
}

func fixtureDocument(size int) (*automerge.Document, error) {
	document, err := automerge.New(benchmarkActor)
	if err != nil {
		return nil, err
	}

	text, err := document.CreateText("body")
	if err != nil {
		_ = document.Close()
		return nil, err
	}

	if err := text.Splice(0, 0, benchmarkText(size)); err != nil {
		_ = document.Close()
		return nil, err
	}

	if _, err := document.Commit("benchmark", time.Time{}); err != nil {
		_ = document.Close()
		return nil, err
	}

	return document, nil
}

func fixtureData(size int, file string) ([]byte, error) {
	if file != "" {
		return os.ReadFile(file)
	}

	document, err := fixtureDocument(size)
	if err != nil {
		return nil, err
	}

	defer func() { _ = document.Close() }()

	return document.Save()
}

func mergeFixtureData(size, tailEdits int) ([]byte, []byte, error) {
	base, err := fixtureDocument(size)
	if err != nil {
		return nil, nil, err
	}

	baseData, err := base.Save()
	_ = base.Close()
	if err != nil {
		return nil, nil, err
	}

	left, err := automerge.Load(baseData, benchmarkActor)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = left.Close() }()

	right, err := automerge.Load(baseData, benchmarkPeerActor)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = right.Close() }()

	leftText, err := left.Text("body")
	if err != nil {
		return nil, nil, err
	}
	for index := range tailEdits {
		if err := leftText.Splice(uint32(size+index), 0, "L"); err != nil {
			return nil, nil, err
		}
	}
	if _, err := left.Commit("benchmark", time.Time{}); err != nil {
		return nil, nil, err
	}

	rightText, err := right.Text("body")
	if err != nil {
		return nil, nil, err
	}
	for index := range tailEdits {
		if err := rightText.Splice(uint32(size+index), 0, "R"); err != nil {
			return nil, nil, err
		}
	}
	if _, err := right.Commit("benchmark", time.Time{}); err != nil {
		return nil, nil, err
	}

	leftData, err := left.Save()
	if err != nil {
		return nil, nil, err
	}

	rightData, err := right.Save()
	if err != nil {
		return nil, nil, err
	}

	return leftData, rightData, nil
}

func benchmarkText(size int) string {
	value := make([]byte, size)
	for index := range value {
		value[index] = byte('a' + index%26)
	}

	return string(value)
}

func mapChecksum(document *automerge.Document, size int) (string, error) {
	values, err := document.Root().Object("values")
	if err != nil {
		return "", err
	}

	normalized := make([]byte, 0, size*16)
	for index := range size {
		value, err := values.Scalar(strconv.Itoa(index))
		if err != nil {
			return "", err
		}

		normalized = strconv.AppendInt(normalized, value.Int, 10)
		normalized = append(normalized, '\n')
	}

	return checksum(normalized), nil
}

func textChecksum(document *automerge.Document) (string, error) {
	text, err := document.Text("body")
	if err != nil {
		return "", err
	}

	value, err := text.String()
	if err != nil {
		return "", err
	}

	return checksum([]byte(value)), nil
}

func synchronize(left, right *automerge.SyncState) error {
	_, err := synchronizeWithWire(left, right)

	return err
}

func synchronizeWithWire(left, right *automerge.SyncState) ([]byte, error) {
	var wire []byte
	for range 100 {
		progressed := false

		message, ok, err := left.GenerateMessage()
		if err != nil {
			return nil, err
		}
		if ok {
			wire = append(wire, message...)
			if err := right.ReceiveMessage(message); err != nil {
				return nil, err
			}
			progressed = true
		}

		message, ok, err = right.GenerateMessage()
		if err != nil {
			return nil, err
		}
		if ok {
			wire = append(wire, message...)
			if err := left.ReceiveMessage(message); err != nil {
				return nil, err
			}
			progressed = true
		}

		if !progressed {
			return wire, nil
		}
	}

	return nil, fmt.Errorf("sync did not quiesce")
}

func checksum(value []byte) string {
	digest := sha256.Sum256(value)

	return hex.EncodeToString(digest[:])
}

func optionalChecksum(value []byte) string {
	if len(value) == 0 {
		return ""
	}

	return checksum(value)
}
