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

package tray

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldStopInteractiveAgentProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		baseName  string
		pid       uint32
		selfPID   uint32
		sessionID uint32
		want      bool
	}{
		{
			name:      "tray in user session",
			baseName:  "probo-agent.exe",
			pid:       42,
			selfPID:   1,
			sessionID: 1,
			want:      true,
		},
		{
			name:      "service in session 0",
			baseName:  "probo-agent.exe",
			pid:       42,
			selfPID:   1,
			sessionID: 0,
			want:      false,
		},
		{
			name:      "skips current process",
			baseName:  "probo-agent.exe",
			pid:       7,
			selfPID:   7,
			sessionID: 1,
			want:      false,
		},
		{
			name:      "other executable",
			baseName:  "notepad.exe",
			pid:       42,
			selfPID:   1,
			sessionID: 1,
			want:      false,
		},
		{
			name:      "case insensitive exe name",
			baseName:  "Probo-Agent.EXE",
			pid:       42,
			selfPID:   1,
			sessionID: 2,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got := shouldStopInteractiveAgentProcess(
					tt.baseName,
					tt.pid,
					tt.selfPID,
					tt.sessionID,
				)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}
