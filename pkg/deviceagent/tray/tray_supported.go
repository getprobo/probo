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

//go:build (darwin && cgo) || windows

package tray

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"fyne.io/systray"
	"go.probo.inc/probo/pkg/deviceagent"
)

func Run(opts Options) error {
	if opts.ServerURL == "" {
		opts.ServerURL = deviceagent.DefaultServerURL
	}

	done := make(chan struct{})

	var shutdown sync.Once

	stop := func() {
		shutdown.Do(func() {
			close(done)
		})
	}

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-sigCh:
			stop()
			systray.Quit()
		case <-done:
		}
	}()

	systray.Run(
		func() {
			onReady(opts, done)
		},
		stop,
	)

	return nil
}

func onReady(opts Options, done <-chan struct{}) {
	setTrayIcons()

	systray.SetTitle("Probo")

	enrollmentRequiredItem := systray.AddMenuItem("Enrollment required", "Enroll from the Probo console")
	enrollmentRequiredItem.Disable()

	connectedItem := systray.AddMenuItem("Connected", "Device is enrolled and reporting")
	connectedItem.Disable()

	statusUnavailableItem := systray.AddMenuItem(
		"Status unavailable",
		"Cannot read enrollment status",
	)
	statusUnavailableItem.Disable()
	statusUnavailableItem.Hide()

	systray.AddSeparator()

	aboutItem := systray.AddMenuItem(
		fmt.Sprintf("About probo-agent %s", opts.Version),
		"Probo device posture agent",
	)

	updateMenu := func() {
		enrolled, err := deviceagent.IsEnrolled(opts.RunDir)
		if err != nil {
			enrollmentRequiredItem.Hide()
			connectedItem.Hide()
			statusUnavailableItem.Show()
			systray.SetTooltip("Probo Device Posture Agent — Status unavailable")

			return
		}

		statusUnavailableItem.Hide()

		if enrolled {
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

		for {
			select {
			case <-ticker.C:
				updateMenu()
			case <-done:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-aboutItem.ClickedCh:
				showAbout(opts.Version)
			case <-done:
				return
			}
		}
	}()
}
