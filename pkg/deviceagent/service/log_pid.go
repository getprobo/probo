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

//go:build !windows

package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

var errStalePIDFile = errors.New("pid file changed while locking")

func writePIDFile(path string) error {
	for range 8 {
		err := writeLockedPID(path)
		if !errors.Is(err, errStalePIDFile) {
			return err
		}
	}

	return fmt.Errorf("cannot write %s: file changed while locking", path)
}

func writeLockedPID(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}

	defer func() { _ = file.Close() }()

	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("cannot lock %s: %w", path, err)
	}

	defer func() { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }()

	if err := pidFileIsLive(path, file); err != nil {
		return err
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("cannot truncate %s: %w", path, err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("cannot rewind %s: %w", path, err)
	}

	body := strconv.Itoa(os.Getpid()) + "\n"
	if _, err := file.Write([]byte(body)); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	if err := file.Chmod(0o644); err != nil {
		return fmt.Errorf("cannot set mode on %s: %w", path, err)
	}

	return nil
}

// removeOwnPIDFile deletes path when it still names this process.
// A crash cannot run this; the next start overwrites the file under
// the same lock. An older process does not delete a newer file.
func removeOwnPIDFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}

	defer func() { _ = file.Close() }()

	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("cannot lock %s: %w", path, err)
	}

	defer func() { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }()

	if err := pidFileIsLive(path, file); err != nil {
		if errors.Is(err, errStalePIDFile) {
			return nil
		}

		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	pid, err := strconv.Atoi(string(bytes.TrimSpace(data)))
	if err != nil || pid != os.Getpid() {
		return nil
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove %s: %w", path, err)
	}

	return nil
}

func pidFileIsLive(path string, file *os.File) error {
	live, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return errStalePIDFile
	}

	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", path, err)
	}

	opened, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat open %s: %w", path, err)
	}

	if !os.SameFile(live, opened) {
		return errStalePIDFile
	}

	return nil
}
