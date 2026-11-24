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

//go:build darwin

package service

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	newsyslogConfPath  = "/etc/newsyslog.d/com.probo.agent.conf"
	darwinAgentLogPath = "/var/log/probo-agent.log"
	darwinAgentPIDPath = "/var/run/probo-agent/probo-agent.pid"
)

//go:embed newsyslog.conf
var newsyslogConf string

var (
	rotationSignals chan os.Signal
	rotationStarted bool
)

// StartAgentLogRotation records the daemon pid and reopens the log when
// newsyslog sends SIGHUP. A non-root process skips this: the LaunchDaemon
// runs as root, and a foreground run must still start.
func StartAgentLogRotation() error {
	if os.Geteuid() != 0 {
		return nil
	}

	if err := writePIDFile(darwinAgentPIDPath); err != nil {
		return fmt.Errorf("cannot write agent pid file: %w", err)
	}

	rotationSignals = make(chan os.Signal, 1)
	signal.Notify(rotationSignals, syscall.SIGHUP)

	rotationStarted = true

	go func() {
		for range rotationSignals {
			if err := reopenAgentLog(); err != nil {
				fmt.Fprintf(os.Stderr, "probo-agent: %v\n", err)
			}
		}
	}()

	return nil
}

// StopAgentLogRotation stops handling SIGHUP and removes the pid file.
func StopAgentLogRotation() {
	if !rotationStarted {
		return
	}

	signal.Stop(rotationSignals)

	rotationStarted = false
	_ = os.Remove(darwinAgentPIDPath)
}

func installNewsyslogConf() error {
	if err := os.MkdirAll(filepath.Dir(newsyslogConfPath), 0o755); err != nil {
		return fmt.Errorf("cannot create newsyslog directory: %w", err)
	}

	if err := os.WriteFile(newsyslogConfPath, []byte(newsyslogConf), 0o644); err != nil {
		return fmt.Errorf("cannot write newsyslog configuration: %w", err)
	}

	return nil
}

func removeNewsyslog() error {
	var errs error
	if err := removeIfPresent(newsyslogConfPath); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := removeIfPresent(darwinAgentPIDPath); err != nil {
		errs = errors.Join(errs, err)
	}

	matches, err := filepath.Glob(darwinAgentLogPath + "*")
	if err != nil {
		return errors.Join(errs, fmt.Errorf("cannot list agent logs: %w", err))
	}

	for _, path := range matches {
		if err := removeIfPresent(path); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func writePIDFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("cannot create agent pid directory: %w", err)
	}

	data := []byte(strconv.Itoa(os.Getpid()) + "\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	return nil
}

func reopenAgentLog() error {
	file, err := os.OpenFile(darwinAgentLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return fmt.Errorf("cannot reopen agent log: %w", err)
	}

	defer func() { _ = file.Close() }()

	fd := int(file.Fd())
	if err := syscall.Dup2(fd, syscall.Stdout); err != nil {
		return fmt.Errorf("cannot replace stdout with the agent log: %w", err)
	}

	if err := syscall.Dup2(fd, syscall.Stderr); err != nil {
		return fmt.Errorf("cannot replace stderr with the agent log: %w", err)
	}

	return nil
}

func removeIfPresent(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove %s: %w", path, err)
	}

	return nil
}
