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

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.probo.inc/probo/pkg/deviceagent"
	"golang.org/x/sys/windows/svc"
)

const (
	windowsRestartExitCode = 75
	windowsFailureExitCode = 1
	windowsStopTimeout     = 10 * time.Second
)

type windowsAgentService struct {
	run func(ctx context.Context) error
	err error
}

func IsWindowsService() (bool, error) {
	isSvc, err := svc.IsWindowsService()
	if err != nil {
		return false, fmt.Errorf("cannot detect windows service: %w", err)
	}

	return isSvc, nil
}

func RunWindowsService(name string, run func(context.Context) error) error {
	h := &windowsAgentService{run: run}
	if err := svc.Run(name, h); err != nil {
		return fmt.Errorf("cannot run windows service: %w", err)
	}

	return h.err
}

func (s *windowsAgentService) Execute(
	_ []string,
	r <-chan svc.ChangeRequest,
	changes chan<- svc.Status,
) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- s.run(ctx)
	}()

	changes <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown,
	}

	var (
		stopping       bool
		stopWait       <-chan struct{}
		stopWaitCancel context.CancelFunc
	)
	defer func() {
		if stopWaitCancel != nil {
			stopWaitCancel()
		}
	}()

	for {
		select {
		case err := <-done:
			s.err = err
			if errors.Is(err, deviceagent.ErrRestartRequired) {
				os.Exit(windowsRestartExitCode)
			}

			changes <- svc.Status{State: svc.StopPending}

			if stopping || err == nil || errors.Is(err, context.Canceled) {
				return false, 0
			}

			return false, windowsFailureExitCode
		case <-stopWait:
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				if stopping {
					break
				}

				stopping = true
				changes <- svc.Status{
					State:    svc.StopPending,
					WaitHint: uint32(windowsStopTimeout / time.Millisecond),
				}
				cancel()

				var waitCtx context.Context
				waitCtx, stopWaitCancel = context.WithTimeout(context.Background(), windowsStopTimeout)
				stopWait = waitCtx.Done()
			}
		}
	}
}
