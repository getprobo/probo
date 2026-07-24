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

package checks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const defaultCommandTimeout = 5 * time.Second

var commandExistsCache sync.Map

// CmdResult captures the basic outcome of an OS subcommand.
type CmdResult struct {
	Stdout string
	Stderr string
	Err    error
}

// RunCommand executes a command and returns trimmed stdout/stderr.
func RunCommand(ctx context.Context, name string, args ...string) CmdResult {
	cmdCtx, cancel := context.WithTimeout(ctx, defaultCommandTimeout)
	defer cancel()

	resolved, ok := resolveCommandPath(name)
	if !ok {
		return CmdResult{
			Err: fmt.Errorf("command %q not available at expected absolute path", name),
		}
	}

	cmd := exec.CommandContext(cmdCtx, resolved, args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return CmdResult{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
		Err:    err,
	}
}

// CommandExists reports whether `cmd` exists at expected absolute path(s).
func CommandExists(cmd string) bool {
	if cached, ok := commandExistsCache.Load(cmd); ok {
		return cached.(bool)
	}

	_, exists := resolveCommandPath(cmd)
	commandExistsCache.Store(cmd, exists)

	return exists
}

// CommandCandidates returns absolute paths for known commands on the current platform.
func CommandCandidates(cmd string) []string {
	return commandCandidates(cmd)
}

func resolveCommandPath(cmd string) (string, bool) {
	if filepath.IsAbs(cmd) {
		return cmd, isExecutableFile(cmd)
	}

	for _, candidate := range commandCandidates(cmd) {
		if isExecutableFile(candidate) {
			return candidate, true
		}
	}

	return "", false
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	if runtime.GOOS == "windows" {
		return true
	}

	return info.Mode().Perm()&0o111 != 0
}
