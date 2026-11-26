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
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	serviceLogPath = "/var/log/probo-agent.log"
	serviceLogMode = 0o644
	serviceLogName = "probo-agent.log"

	newsyslogCount  = 7
	newsyslogSizeKB = 1024

	darwinNewsyslogPath = "/etc/newsyslog.d/probo-agent.conf"
	darwinPIDPath       = "/var/run/probo-agent.pid"

	freebsdNewsyslogPath = "/usr/local/etc/newsyslog.conf.d/probo-agent.conf"
	freebsdSyslogPath    = "/usr/local/etc/syslog.d/probo-agent.conf"
	freebsdSyslogPIDPath = "/var/run/syslog.pid"
)

// reopenFile appends to a log file and can open the path again after
// newsyslog renames it.
type reopenFile struct {
	path   string
	mu     sync.Mutex
	file   *os.File
	closed bool
	stop   chan struct{}
	once   sync.Once
}

func openLogFile(path string) (*reopenFile, error) {
	file, err := openAppend(path)
	if err != nil {
		return nil, err
	}

	return &reopenFile{
		path: path,
		file: file,
		stop: make(chan struct{}),
	}, nil
}

func openAppend(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, serviceLogMode)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}

	if err := file.Chmod(serviceLogMode); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("cannot set mode on %s: %w", path, err)
	}

	return file, nil
}

func (r *reopenFile) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		return 0, os.ErrClosed
	}

	return r.file.Write(p)
}

func (r *reopenFile) reopen() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()

		return os.ErrClosed
	}

	r.mu.Unlock()

	file, err := openAppend(r.path)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		_ = file.Close()

		return os.ErrClosed
	}

	old := r.file
	r.file = file

	if old != nil {
		_ = old.Close()
	}

	return nil
}

func (r *reopenFile) Close() error {
	r.mu.Lock()
	r.closed = true
	file := r.file
	r.file = nil
	r.mu.Unlock()

	r.once.Do(func() {
		close(r.stop)
	})

	if file == nil {
		return nil
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("cannot close %s: %w", r.path, err)
	}

	return nil
}

// watchHUP reopens the log when signals receives SIGHUP. A closed stop
// wins over a signal already sitting in the buffer, so shutdown does not
// create the file again.
func (r *reopenFile) watchHUP(signals <-chan os.Signal, done chan<- struct{}) {
	defer close(done)

	for {
		select {
		case <-r.stop:
			return
		case <-signals:
			select {
			case <-r.stop:
				return
			default:
				_ = r.reopen()
			}
		}
	}
}

func newsyslogEntry(logPath, pidPath string) string {
	return fmt.Sprintf(
		"%s\troot:wheel\t644\t%d\t%d\t*\tZC\t%s\t1\n",
		logPath,
		newsyslogCount,
		newsyslogSizeKB,
		pidPath,
	)
}

func darwinNewsyslogConf() string {
	return newsyslogEntry(serviceLogPath, darwinPIDPath)
}

func freebsdNewsyslogConf() string {
	return newsyslogEntry(serviceLogPath, freebsdSyslogPIDPath)
}

func freebsdSyslogConf() string {
	return "!probo-agent\n*.*\t" + serviceLogPath + "\n!*\n"
}

func writeRootFile(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("cannot create %s: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	if err := os.Chmod(path, 0o644); err != nil {
		return fmt.Errorf("cannot set mode on %s: %w", path, err)
	}

	return nil
}

func removeFile(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove %s: %w", path, err)
	}

	return nil
}

func removePrefixedLogs(dir, name string) error {
	matches, err := filepath.Glob(filepath.Join(dir, name+"*"))
	if err != nil {
		return fmt.Errorf("cannot list logs in %s: %w", dir, err)
	}

	for _, match := range matches {
		if err := os.Remove(match); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("cannot remove %s: %w", match, err)
		}
	}

	return nil
}

func logEventLevel(line string) string {
	switch {
	case strings.Contains(line, `"level":"ERROR"`):
		return "error"
	case strings.Contains(line, `"level":"WARN"`):
		return "warning"
	default:
		return "info"
	}
}
