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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.probo.inc/probo/pkg/automerge"
)

type result struct {
	Workload   string `json:"workload"`
	Size       int    `json:"size"`
	Iterations int    `json:"iterations"`
	TotalNS    int64  `json:"totalNs"`
	NSPerOp    int64  `json:"nsPerOp"`
	Checksum   string `json:"checksum"`
}

type benchmarkWorkload struct {
	run      func() error
	validate func() (string, error)
	cleanup  func()
}

var benchmarkActor = automerge.ActorID{
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 11, 12, 13, 14, 15,
}

func main() {
	workload := flag.String("workload", "", "benchmark workload")
	size := flag.Int("size", 0, "workload size")
	iterations := flag.Int("iterations", 1, "timed iterations")
	warmups := flag.Int("warmups", 3, "warmup iterations")
	fixture := flag.String("fixture", "", "shared fixture path")

	flag.Parse()

	if *workload == "" || *iterations <= 0 || *warmups < 0 {
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

		data, err := document.Save(context.Background())
		_ = document.Close(context.Background())

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

	startedAt := time.Now()

	for range *iterations {
		if err := runner.run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	total := time.Since(startedAt)

	checksum, err := runner.validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := json.NewEncoder(os.Stdout).Encode(result{
		Workload:   *workload,
		Size:       *size,
		Iterations: *iterations,
		TotalNS:    total.Nanoseconds(),
		NSPerOp:    total.Nanoseconds() / int64(*iterations),
		Checksum:   checksum,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func workloadRunner(
	workload string,
	size int,
	fixture string,
) (benchmarkWorkload, error) {
	switch workload {
	case "create":
		return benchmarkWorkload{
			run: func() error {
				document, err := automerge.New(
					context.Background(),
					benchmarkActor,
				)
				if err != nil {
					return err
				}

				return document.Close(context.Background())
			},
			validate: func() (string, error) {
				return checksum([]byte("empty")), nil
			},
			cleanup: func() {},
		}, nil
	case "map":
		return benchmarkWorkload{
			run: func() error {
				document, err := mapDocument(size)
				if err != nil {
					return err
				}

				return document.Close(context.Background())
			},
			validate: func() (string, error) {
				document, err := mapDocument(size)
				if err != nil {
					return "", err
				}

				defer func() { _ = document.Close(context.Background()) }()

				return mapChecksum(document, size)
			},
			cleanup: func() {},
		}, nil
	case "text":
		return benchmarkWorkload{
			run: func() error {
				document, err := typedDocument(size)
				if err != nil {
					return err
				}

				return document.Close(context.Background())
			},
			validate: func() (string, error) {
				document, err := typedDocument(size)
				if err != nil {
					return "", err
				}

				defer func() { _ = document.Close(context.Background()) }()

				return textChecksum(document)
			},
			cleanup: func() {},
		}, nil
	case "load":
		data, err := fixtureData(size, fixture)
		if err != nil {
			return benchmarkWorkload{}, err
		}

		return benchmarkWorkload{
			run: func() error {
				loaded, err := automerge.Load(
					context.Background(),
					data,
					benchmarkActor,
				)
				if err != nil {
					return err
				}

				return loaded.Close(context.Background())
			},
			validate: func() (string, error) {
				loaded, err := automerge.Load(
					context.Background(),
					data,
					benchmarkActor,
				)
				if err != nil {
					return "", err
				}

				defer func() { _ = loaded.Close(context.Background()) }()

				return textChecksum(loaded)
			},
			cleanup: func() {},
		}, nil
	case "save":
		data, err := fixtureData(size, fixture)
		if err != nil {
			return benchmarkWorkload{}, err
		}

		document, err := automerge.Load(
			context.Background(),
			data,
			benchmarkActor,
		)
		if err != nil {
			return benchmarkWorkload{}, err
		}

		return benchmarkWorkload{
			run: func() error {
				_, err := document.Save(context.Background())

				return err
			},
			validate: func() (string, error) {
				data, err := document.Save(context.Background())
				if err != nil {
					return "", err
				}

				loaded, err := automerge.Load(
					context.Background(),
					data,
					benchmarkActor,
				)
				if err != nil {
					return "", err
				}

				defer func() { _ = loaded.Close(context.Background()) }()

				return textChecksum(loaded)
			},
			cleanup: func() {
				_ = document.Close(context.Background())
			},
		}, nil
	default:
		return benchmarkWorkload{}, fmt.Errorf("unknown workload %q", workload)
	}
}

func mapDocument(size int) (*automerge.Document, error) {
	ctx := context.Background()

	document, err := automerge.New(ctx, benchmarkActor)
	if err != nil {
		return nil, err
	}

	values, err := document.Root().CreateObject(
		ctx,
		"values",
		automerge.ObjectTypeMap,
	)
	if err != nil {
		_ = document.Close(ctx)
		return nil, err
	}

	for index := range size {
		if err := values.PutScalar(
			ctx,
			strconv.Itoa(index),
			automerge.Scalar{
				Type: automerge.ScalarTypeInt,
				Int:  int64(index),
			},
		); err != nil {
			_ = document.Close(ctx)
			return nil, err
		}
	}

	if _, err := document.Commit(ctx, "benchmark", time.Time{}); err != nil {
		_ = document.Close(ctx)
		return nil, err
	}

	return document, nil
}

func typedDocument(size int) (*automerge.Document, error) {
	ctx := context.Background()

	document, err := automerge.New(ctx, benchmarkActor)
	if err != nil {
		return nil, err
	}

	text, err := document.CreateText(ctx, "body")
	if err != nil {
		_ = document.Close(ctx)
		return nil, err
	}

	for index := range size {
		if err := text.Splice(ctx, uint32(index), 0, "x"); err != nil {
			_ = document.Close(ctx)
			return nil, err
		}
	}

	if _, err := document.Commit(ctx, "benchmark", time.Time{}); err != nil {
		_ = document.Close(ctx)
		return nil, err
	}

	return document, nil
}

func fixtureDocument(size int) (*automerge.Document, error) {
	ctx := context.Background()

	document, err := automerge.New(ctx, benchmarkActor)
	if err != nil {
		return nil, err
	}

	text, err := document.CreateText(ctx, "body")
	if err != nil {
		_ = document.Close(ctx)
		return nil, err
	}

	if err := text.Splice(ctx, 0, 0, benchmarkText(size)); err != nil {
		_ = document.Close(ctx)
		return nil, err
	}

	if _, err := document.Commit(ctx, "benchmark", time.Time{}); err != nil {
		_ = document.Close(ctx)
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

	defer func() { _ = document.Close(context.Background()) }()

	return document.Save(context.Background())
}

func benchmarkText(size int) string {
	value := make([]byte, size)
	for index := range value {
		value[index] = byte('a' + index%26)
	}

	return string(value)
}

func mapChecksum(document *automerge.Document, size int) (string, error) {
	values, err := document.Root().Object(context.Background(), "values")
	if err != nil {
		return "", err
	}

	normalized := make([]byte, 0, size*16)
	for index := range size {
		value, err := values.Scalar(
			context.Background(),
			strconv.Itoa(index),
		)
		if err != nil {
			return "", err
		}

		normalized = strconv.AppendInt(normalized, value.Int, 10)
		normalized = append(normalized, '\n')
	}

	return checksum(normalized), nil
}

func textChecksum(document *automerge.Document) (string, error) {
	text, err := document.Text(context.Background(), "body")
	if err != nil {
		return "", err
	}

	value, err := text.String(context.Background())
	if err != nil {
		return "", err
	}

	return checksum([]byte(value)), nil
}

func checksum(value []byte) string {
	digest := sha256.Sum256(value)

	return hex.EncodeToString(digest[:])
}
