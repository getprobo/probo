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
	"fmt"
	"io"
	"os"
	"strings"
)

// WriteDiagnosticsOutput writes the report to a named pipe. A regular
// file path is refused so an elevated run cannot replace an arbitrary file.
func WriteDiagnosticsOutput(path, report string, _ int) error {
	if path == "" {
		return fmt.Errorf("diagnostics output path is empty")
	}

	if !strings.HasPrefix(path, `\\.\pipe\`) {
		return fmt.Errorf("diagnostics output must be a named pipe: %s", path)
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("cannot open diagnostics output: %w", err)
	}

	_, err = io.WriteString(file, report)
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("cannot write diagnostics report: %w", err)
	}

	if closeErr != nil {
		return fmt.Errorf("cannot close diagnostics output: %w", closeErr)
	}

	return nil
}
