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
	"fmt"
	"io"
	"log/syslog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OpenRunLog sends service logs to syslogd, which writes
// /var/log/probo-agent.log. newsyslog signals syslogd, not this process.
func OpenRunLog() (io.WriteCloser, error) {
	writer, err := syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, "probo-agent")
	if err != nil {
		return nil, fmt.Errorf("cannot open syslog: %w", err)
	}

	return writer, nil
}

func installServiceLog() error {
	if err := writeRootFile(freebsdSyslogPath, freebsdSyslogConf()); err != nil {
		return err
	}

	if err := writeRootFile(freebsdNewsyslogPath, freebsdNewsyslogConf()); err != nil {
		return err
	}

	if err := ensureServiceLogFile(); err != nil {
		return err
	}

	return reloadSyslogd()
}

func uninstallServiceLog() error {
	if err := removeFile(freebsdSyslogPath); err != nil {
		return err
	}

	if err := removeFile(freebsdNewsyslogPath); err != nil {
		return err
	}

	if err := removePrefixedLogs(filepath.Dir(serviceLogPath), serviceLogName); err != nil {
		return err
	}

	return reloadSyslogd()
}

func ensureServiceLogFile() error {
	file, err := os.OpenFile(serviceLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, serviceLogMode)
	if err != nil {
		return fmt.Errorf("cannot create service log: %w", err)
	}

	defer func() { _ = file.Close() }()

	if err := file.Chmod(serviceLogMode); err != nil {
		return fmt.Errorf("cannot set service log mode: %w", err)
	}

	return nil
}

func reloadSyslogd() error {
	out, err := exec.Command("service", "syslogd", "reload").CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot reload syslogd: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}
