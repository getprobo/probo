// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, publish, distribute, sublicense, and/or sell
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
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"fyne.io/systray"
	"go.probo.inc/probo/pkg/deviceagent"
)

func Run(opts Options) error {
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

type enrollmentMenu struct {
	items []*systray.MenuItem
}

func (m enrollmentMenu) show() {
	for _, item := range m.items {
		item.Show()
	}
}

func (m enrollmentMenu) hide() {
	for _, item := range m.items {
		item.Hide()
	}
}

func onReady(opts Options, done <-chan struct{}) {
	setTrayIcons()
	setTrayTitle()

	connectedItem := systray.AddMenuItem("Connected", "Device is enrolled and reporting")
	connectedItem.SetIcon(statusConnectedIconData)
	connectedItem.Disable()
	connectedItem.Hide()

	enrollmentRequiredItem := systray.AddMenuItem(
		"Enrollment required",
		"Device is not enrolled yet",
	)
	enrollmentRequiredItem.SetIcon(statusEnrollmentIconData)
	enrollmentRequiredItem.Disable()
	enrollmentRequiredItem.Hide()

	statusUnavailableItem := systray.AddMenuItem(
		"Status unavailable",
		"Cannot read enrollment status",
	)
	statusUnavailableItem.SetIcon(statusUnavailableIconData)
	statusUnavailableItem.Disable()
	statusUnavailableItem.Hide()

	systray.AddSeparator()

	enrollMenu := setupEnrollmentMenu(opts, done)

	aboutItem := systray.AddMenuItem(
		"About Probo Device Posture Agent…",
		"Probo device posture agent",
	)

	updateMenu := func() {
		enrolled, err := deviceagent.IsEnrolled(opts.RunDir)
		if err != nil {
			enrollMenu.hide()
			connectedItem.Hide()
			enrollmentRequiredItem.Hide()
			statusUnavailableItem.Show()
			systray.SetTooltip("Probo Device Posture Agent — Status unavailable")

			return
		}

		statusUnavailableItem.Hide()

		if enrolled {
			enrollMenu.hide()
			enrollmentRequiredItem.Hide()
			connectedItem.Show()
			systray.SetTooltip("Probo Device Posture Agent — Connected")
		} else {
			enrollMenu.show()
			connectedItem.Hide()
			enrollmentRequiredItem.Show()
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

func setupEnrollmentMenu(opts Options, done <-chan struct{}) enrollmentMenu {
	const enrollTooltip = "Open enrollment in your browser"

	if opts.ServerURL != "" {
		enrollItem := systray.AddMenuItem("Enroll in browser", enrollTooltip)
		bindMenuClick(enrollItem, done, func() {
			openConsoleEnroll(opts.ServerURL)
		})

		return enrollmentMenu{items: []*systray.MenuItem{enrollItem}}
	}

	enrollRoot := systray.AddMenuItem("Enroll via…", enrollTooltip)

	usItem := enrollRoot.AddSubMenuItem(
		"United States (us.probo.com)",
		"Open US console enrollment",
	)
	bindMenuClick(usItem, done, func() {
		openConsoleEnroll(deviceagent.USConsoleURL)
	})

	euItem := enrollRoot.AddSubMenuItem(
		"European Union (eu.probo.com)",
		"Open EU console enrollment",
	)
	bindMenuClick(euItem, done, func() {
		openConsoleEnroll(deviceagent.EUConsoleURL)
	})

	selfHostedItem := enrollRoot.AddSubMenuItem(
		"Self hosted…",
		"Enter your Probo hostname",
	)
	bindMenuClick(selfHostedItem, done, openSelfHostedEnroll)

	return enrollmentMenu{items: []*systray.MenuItem{enrollRoot}}
}

func bindMenuClick(item *systray.MenuItem, done <-chan struct{}, fn func()) {
	go func() {
		for {
			select {
			case <-item.ClickedCh:
				fn()
			case <-done:
				return
			}
		}
	}()
}
