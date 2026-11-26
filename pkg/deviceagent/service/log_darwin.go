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
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// hupLog reopens the log on SIGHUP and drops this process's PID on close.
type hupLog struct {
	file    *reopenFile
	signals chan os.Signal
	done    chan struct{}
}

// OpenRunLog opens the service log. A non-root run cannot create
// /var/log/probo-agent.log and the caller keeps stderr.
func OpenRunLog() (io.WriteCloser, error) {
	file, err := openLogFile(serviceLogPath)
	if err != nil {
		return nil, err
	}

	// Install the handler before the PID file exists. newsyslog signals
	// that PID, and SIGHUP's default action is to terminate.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)

	if err := writePIDFile(darwinPIDPath); err != nil {
		signal.Stop(signals)

		_ = file.Close()

		return nil, err
	}

	done := make(chan struct{})
	go file.watchHUP(signals, done)

	return &hupLog{file: file, signals: signals, done: done}, nil
}

func (h *hupLog) Write(p []byte) (int, error) {
	return h.file.Write(p)
}

func (h *hupLog) Close() error {
	err := removeOwnPIDFile(darwinPIDPath)

	signal.Stop(h.signals)

	if closeErr := h.file.Close(); err == nil {
		err = closeErr
	}

	<-h.done

	return err
}

func installServiceLog() error {
	return writeRootFile(darwinNewsyslogPath, darwinNewsyslogConf())
}

func uninstallServiceLog() error {
	if err := removeFile(darwinNewsyslogPath); err != nil {
		return err
	}

	if err := removeFile(darwinPIDPath); err != nil {
		return err
	}

	return removePrefixedLogs(filepath.Dir(serviceLogPath), serviceLogName)
}
