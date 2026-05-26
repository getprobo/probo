// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
