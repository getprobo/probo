// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

//go:build (darwin || windows) && cgo

package tray

import (
	"fmt"
	"time"

	"fyne.io/systray"
	"go.probo.inc/probo/pkg/deviceagent"
)

func Run(opts Options) error {
	if opts.ServerURL == "" {
		opts.ServerURL = deviceagent.DefaultServerURL
	}

	systray.Run(
		func() {
			onReady(opts)
		},
		func() {},
	)

	return nil
}

func onReady(opts Options) {
	setTrayIcons()

	systray.SetTitle("Probo")

	enrollItem := systray.AddMenuItem(
		"Provide enrollment token…",
		"Enroll this device with your Probo workspace",
	)
	connectedItem := systray.AddMenuItem("Connected", "Device is enrolled and reporting")
	connectedItem.Disable()
	systray.AddSeparator()
	aboutItem := systray.AddMenuItem(
		fmt.Sprintf("About probo-agent %s", opts.Version),
		"Probo device posture agent",
	)
	quitItem := systray.AddMenuItem("Quit", "Quit the menu bar helper")

	updateMenu := func() {
		if deviceagent.IsEnrolled(opts.Dir) {
			enrollItem.Hide()
			connectedItem.Show()
			systray.SetTooltip("Probo Device Posture Agent — Connected")
		} else {
			enrollItem.Show()
			connectedItem.Hide()
			systray.SetTooltip("Probo Device Posture Agent — Enrollment required")
		}
	}

	updateMenu()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			updateMenu()
		}
	}()

	if opts.PromptEnrollment && !deviceagent.IsEnrolled(opts.Dir) {
		go func() {
			// Give Installer.app time to close before showing dialogs.
			time.Sleep(1 * time.Second)
			handleEnrollment(opts, updateMenu)
		}()
	}

	go func() {
		for {
			select {
			case <-enrollItem.ClickedCh:
				handleEnrollment(opts, updateMenu)
			case <-aboutItem.ClickedCh:
				showAbout(opts.Version)
			case <-quitItem.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func handleEnrollment(opts Options, updateMenu func()) {
	serverURL, token, ok := promptEnrollment(opts.ExePath)
	if !ok {
		return
	}

	enrollOpts := opts
	enrollOpts.ServerURL = serverURL

	if err := runElevatedInstall(enrollOpts, token); err != nil {
		showError("Enrollment failed", err)
		return
	}

	updateMenu()
}
