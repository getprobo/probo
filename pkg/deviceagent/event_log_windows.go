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

//go:build windows

package deviceagent

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/eventlog"
)

type windowsEventLog struct {
	*eventLogWriter
	source *eventlog.Log
}

// OpenWindowsEventLog opens the ProboAgent source in the Application log.
func OpenWindowsEventLog() (io.WriteCloser, error) {
	source, err := eventlog.Open(WindowsEventSource)
	if err != nil {
		return nil, fmt.Errorf("cannot open Windows event log: %w", err)
	}

	return &windowsEventLog{
		eventLogWriter: newEventLogWriter(source),
		source:         source,
	}, nil
}

func (l *windowsEventLog) Close() error {
	flushErr := l.eventLogWriter.Close()

	sourceErr := l.source.Close()
	if sourceErr != nil {
		sourceErr = fmt.Errorf("cannot close Windows event log: %w", sourceErr)
	}

	return errors.Join(flushErr, sourceErr)
}

// InstallWindowsEventSource registers ProboAgent in the Application log.
// An existing source is left in place so a second install can continue.
func InstallWindowsEventSource() error {
	err := eventlog.InstallAsEventCreate(
		WindowsEventSource,
		eventlog.Info|eventlog.Warning|eventlog.Error,
	)
	if err == nil || strings.Contains(err.Error(), "already exists") {
		return nil
	}

	return fmt.Errorf("cannot register Windows event source: %w", err)
}

// RemoveWindowsEventSource removes the ProboAgent source. A missing
// source is success.
func RemoveWindowsEventSource() error {
	err := eventlog.Remove(WindowsEventSource)
	if err == nil || isMissingEventSource(err) {
		return nil
	}

	return fmt.Errorf("cannot remove Windows event source: %w", err)
}

func isMissingEventSource(err error) bool {
	return errors.Is(err, registry.ErrNotExist) ||
		errors.Is(err, windows.ERROR_FILE_NOT_FOUND) ||
		errors.Is(err, windows.ERROR_PATH_NOT_FOUND)
}
