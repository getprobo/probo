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
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReopenFile_WritesToTheNewInode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, serviceLogName)
	file, err := openLogFile(path)
	require.NoError(t, err)

	_, err = file.Write([]byte("one\n"))
	require.NoError(t, err)

	rotated := path + ".0"
	require.NoError(t, os.Rename(path, rotated))
	require.NoError(t, file.reopen())

	_, err = file.Write([]byte("two\n"))
	require.NoError(t, err)
	require.NoError(t, file.Close())

	old, err := os.ReadFile(rotated)
	require.NoError(t, err)
	assert.Equal(t, "one\n", string(old))

	fresh, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "two\n", string(fresh))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(serviceLogMode), info.Mode().Perm())
}

func TestWatchHUP_StopWinsOverQueuedSignal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, serviceLogName)
	file, err := openLogFile(path)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	require.NoError(t, os.Remove(path))

	signals := make(chan os.Signal, 1)
	signals <- syscall.SIGHUP

	done := make(chan struct{})
	go file.watchHUP(signals, done)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop")
	}

	_, err = os.Stat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
}
