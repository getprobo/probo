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

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc"
)

func TestWindowsAgentService_Execute_ReportsRunningThenStops(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	runCtxDone := make(chan struct{})

	h := &windowsAgentService{
		run: func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			close(runCtxDone)

			return ctx.Err()
		},
	}

	reqs := make(chan svc.ChangeRequest, 1)
	statuses := make(chan svc.Status, 8)

	done := make(chan struct{})
	var errno uint32

	go func() {
		_, errno = h.Execute(nil, reqs, statuses)
		close(done)
	}()

	st := recvStatus(t, statuses)
	assert.Equal(t, svc.StartPending, st.State)

	st = recvStatus(t, statuses)
	assert.Equal(t, svc.Running, st.State)
	assert.Equal(t, svc.AcceptStop|svc.AcceptShutdown, st.Accepts)

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("run was not started")
	}

	reqs <- svc.ChangeRequest{Cmd: svc.Stop}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Execute did not return after stop")
	}

	select {
	case <-runCtxDone:
	case <-time.After(2 * time.Second):
		t.Fatal("run context was not canceled")
	}

	st = recvStatus(t, statuses)
	assert.Equal(t, svc.StopPending, st.State)
	assert.Equal(t, uint32(0), errno)
}

func recvStatus(t *testing.T, statuses <-chan svc.Status) svc.Status {
	t.Helper()

	select {
	case st := <-statuses:
		return st
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for service status")
		return svc.Status{}
	}
}

func TestWindowsAgentService_Execute_PreservesRunError(t *testing.T) {
	t.Parallel()

	want := context.DeadlineExceeded
	h := &windowsAgentService{
		run: func(ctx context.Context) error {
			return want
		},
	}

	reqs := make(chan svc.ChangeRequest)
	statuses := make(chan svc.Status, 8)

	_, errno := h.Execute(nil, reqs, statuses)
	require.Equal(t, uint32(windowsFailureExitCode), errno)
	assert.ErrorIs(t, h.err, want)
}

func TestWindowsAgentService_Execute_ReportsZeroExitOnSuccess(t *testing.T) {
	t.Parallel()

	h := &windowsAgentService{
		run: func(ctx context.Context) error {
			return nil
		},
	}

	reqs := make(chan svc.ChangeRequest)
	statuses := make(chan svc.Status, 8)

	_, errno := h.Execute(nil, reqs, statuses)
	require.Equal(t, uint32(0), errno)
	assert.NoError(t, h.err)
}
