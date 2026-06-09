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

	enrollmentRequiredItem := systray.AddMenuItem("Enrollment required", "Enroll from the Probo console")
	enrollmentRequiredItem.Disable()
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
			enrollmentRequiredItem.Hide()
			connectedItem.Show()
			systray.SetTooltip("Probo Device Posture Agent — Connected")
		} else {
			enrollmentRequiredItem.Show()
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

	go func() {
		for {
			select {
			case <-aboutItem.ClickedCh:
				showAbout(opts.Version)
			case <-quitItem.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}
