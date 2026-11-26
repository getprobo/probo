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

package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/eventlog"
)

const eventID = 1

type eventWriter struct {
	mu  sync.Mutex
	log *eventlog.Log
}

// OpenRunLog writes the Application event log when this process is the
// Windows service. A console run returns nil and stays on stderr.
func OpenRunLog() (io.WriteCloser, error) {
	isSvc, err := IsWindowsService()
	if err != nil || !isSvc {
		return nil, nil
	}

	lg, err := eventlog.Open(DefaultWindowsName)
	if err != nil {
		return nil, fmt.Errorf("cannot open event log: %w", err)
	}

	return &eventWriter{log: lg}, nil
}

func installServiceLog() error {
	err := eventlog.Remove(DefaultWindowsName)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("cannot reset event source: %w", err)
	}

	err = eventlog.InstallAsEventCreate(
		DefaultWindowsName,
		eventlog.Error|eventlog.Warning|eventlog.Info,
	)
	if err != nil {
		return fmt.Errorf("cannot register event source: %w", err)
	}

	return nil
}

// uninstallServiceLog drops the source registration. Application events
// already written stay in the log.
func uninstallServiceLog() error {
	err := eventlog.Remove(DefaultWindowsName)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("cannot remove event source: %w", err)
	}

	return nil
}

func (w *eventWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\r\n")

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.log == nil {
		return 0, fmt.Errorf("cannot write closed event log: %w", os.ErrClosed)
	}

	var err error

	switch logEventLevel(msg) {
	case "error":
		err = w.log.Error(eventID, msg)
	case "warning":
		err = w.log.Warning(eventID, msg)
	default:
		err = w.log.Info(eventID, msg)
	}

	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *eventWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.log == nil {
		return nil
	}

	err := w.log.Close()
	w.log = nil

	if err != nil {
		return fmt.Errorf("cannot close event log: %w", err)
	}

	return nil
}
