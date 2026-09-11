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
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath   = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValueName = "ProboAgentTray"
)

func RegisterAutoStart(exePath string, runDir string) error {
	guiExePath, err := registerAutoStart(exePath, runDir)
	if err != nil {
		return err
	}

	startTrayBestEffort(guiExePath, runDir)

	return nil
}

// RefreshAutoStart updates the Run entry without starting another tray process.
// NOTE: Remove this function with the Run-entry migration after all
// supported installs register probo-agentw.exe.
func RefreshAutoStart(exePath string, runDir string) error {
	_, err := registerAutoStart(exePath, runDir)

	return err
}

func registerAutoStart(exePath string, runDir string) (string, error) {
	if exePath == "" {
		return "", fmt.Errorf("executable path is required")
	}

	if runDir == "" {
		return "", fmt.Errorf("enrollment run directory is required")
	}

	guiExePath := guiExecutablePath(exePath)
	if _, err := os.Stat(guiExePath); err != nil {
		return "", fmt.Errorf("cannot access GUI executable %s: %w", guiExePath, err)
	}

	// HKLM Run is the machine-wide equivalent of /Library/LaunchAgents:
	// every interactive user gets the tray at logon, including after an
	// MSI install that had no GUI session yet.
	command := trayRunCommand(guiExePath, runDir)

	if err := setMachineRunValue(command); err != nil {
		return "", err
	}

	return guiExePath, nil
}

func trayRunCommand(exePath string, runDir string) string {
	return fmt.Sprintf(`"%s" tray --run-dir "%s"`, exePath, runDir)
}

func guiExecutablePath(exePath string) string {
	name := "probo-agentw"
	if filepath.Ext(exePath) != "" {
		name = agentGUIExeBaseName
	}

	return filepath.Join(filepath.Dir(exePath), name)
}

func UnregisterAutoStart() error {
	err := deleteMachineRunValue()
	stopInteractiveAgentProcessesBestEffort()
	return err
}

func setMachineRunValue(command string) error {
	key, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		runKeyPath,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("cannot open or create machine Run registry key: %w", err)
	}

	defer func() { _ = key.Close() }()

	existing, _, err := key.GetStringValue(runValueName)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("cannot read Run registry value: %w", err)
	}

	if err != nil || existing != command {
		if err := key.SetStringValue(runValueName, command); err != nil {
			return fmt.Errorf("cannot set Run registry value: %w", err)
		}
	}

	return nil
}

func deleteMachineRunValue() error {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		runKeyPath,
		registry.SET_VALUE,
	)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("cannot open machine Run registry key: %w", err)
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
