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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInteractiveUserSID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sid  string
		want bool
	}{
		{
			name: "domain user sid",
			sid:  "S-1-5-21-1234567890-123456789-123456789-1001",
			want: true,
		},
		{
			name: "entra user sid",
			sid:  "S-1-12-1-3603547745-1252762009-756918658-301435180",
			want: true,
		},
		{
			name: "entra authority without user subauthority",
			sid:  "S-1-12-2-3603547745-1252762009-756918658-301435180",
			want: false,
		},
		{
			name: "bare entra authority prefix",
			sid:  "S-1-12-",
			want: false,
		},
		{
			name: "local system sid",
			sid:  "S-1-5-18",
			want: false,
		},
		{
			name: "builtin administrators sid",
			sid:  "S-1-5-32-544",
			want: false,
		},
		{
			name: "empty sid",
			sid:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, isInteractiveUserSID(tt.sid))
			},
		)
	}
}

func TestAppendUniqueSessionID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ids       []uint32
		sessionID uint32
		want      []uint32
	}{
		{
			name:      "append to empty slice",
			ids:       nil,
			sessionID: 2,
			want:      []uint32{2},
		},
		{
			name:      "append new session",
			ids:       []uint32{1},
			sessionID: 2,
			want:      []uint32{1, 2},
		},
		{
			name:      "skip duplicate session",
			ids:       []uint32{1, 2},
			sessionID: 2,
			want:      []uint32{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got := appendUniqueSessionID(tt.ids, tt.sessionID)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}
