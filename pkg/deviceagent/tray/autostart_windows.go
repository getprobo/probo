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

package tray

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

const runValueName = "ProboAgentTray"

func RegisterAutoStart(exePath string, runDir string) error {
	if exePath == "" {
		return fmt.Errorf("executable path is required")
	}
	if runDir == "" {
		return fmt.Errorf("enrollment run directory is required")
	}

	sid, err := currentInteractiveUserSID()
	if err != nil {
		return fmt.Errorf("cannot resolve interactive user for tray auto-start: %w", err)
	}

	keyPath := sid + `\` + runKeyPath

	key, _, err := registry.CreateKey(registry.USERS, keyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("cannot open or create Run registry key for interactive user: %w", err)
	}

	defer func() { _ = key.Close() }()

	command := fmt.Sprintf(`"%s" tray --run-dir "%s"`, exePath, runDir)
	if err := key.SetStringValue(runValueName, command); err != nil {
		return fmt.Errorf("cannot set Run registry value: %w", err)
	}

	return nil
}

func UnregisterAutoStart() error {
	sid, err := currentInteractiveUserSID()
	if err != nil {
		return fmt.Errorf("cannot resolve interactive user for tray auto-start: %w", err)
	}

	keyPath := sid + `\` + runKeyPath

	key, err := registry.OpenKey(registry.USERS, keyPath, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("cannot open Run registry key for interactive user: %w", err)
	}

	defer func() { _ = key.Close() }()

	if err := key.DeleteValue(runValueName); err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("cannot delete Run registry value: %w", err)
	}

	return nil
}
