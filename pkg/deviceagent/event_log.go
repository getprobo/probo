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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const (
	// WindowsEventSource is the Application log source used by the service.
	WindowsEventSource = "ProboAgent"

	windowsEventMessageID = 1
)

type eventReporter interface {
	Info(eid uint32, msg string) error
	Warning(eid uint32, msg string) error
	Error(eid uint32, msg string) error
}

// eventLogWriter reports each complete logger line to the Windows event log.
type eventLogWriter struct {
	mu       sync.Mutex
	reporter eventReporter
	buf      []byte
}

func newEventLogWriter(reporter eventReporter) *eventLogWriter {
	return &eventLogWriter{reporter: reporter}
}

func (w *eventLogWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}

		line := bytes.TrimRight(w.buf[:i], "\r")
		w.buf = append([]byte(nil), w.buf[i+1:]...)

		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		if err := reportEventLine(w.reporter, line); err != nil {
			return len(p), err
		}
	}

	return len(p), nil
}

func (w *eventLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	line := bytes.TrimSpace(w.buf)
	w.buf = nil

	if len(line) == 0 {
		return nil
	}

	return reportEventLine(w.reporter, line)
}

func reportEventLine(reporter eventReporter, line []byte) error {
	var rec struct {
		Level string `json:"level"`
	}

	_ = json.Unmarshal(line, &rec)

	msg := string(line)

	var err error

	switch strings.ToUpper(rec.Level) {
	case "ERROR":
		err = reporter.Error(windowsEventMessageID, msg)
	case "WARN", "WARNING":
		err = reporter.Warning(windowsEventMessageID, msg)
	default:
		err = reporter.Info(windowsEventMessageID, msg)
	}

	if err != nil {
		return fmt.Errorf("cannot write Windows event log: %w", err)
	}

	return nil
}
