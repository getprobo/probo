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

package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinuxPasswordPolicyFromDefs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		minLen string
		want   Status
	}{
		{name: "short length is a reading", minLen: "4", want: StatusPass},
		{name: "long length is a reading", minLen: "12", want: StatusPass},
		{name: "zero length is a reading", minLen: "0", want: StatusPass},
		{name: "unset length fails", minLen: "", want: StatusFail},
		{name: "unreadable length is unknown", minLen: "nope", want: StatusUnknown},
		{name: "negative length is unknown", minLen: "-1", want: StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, linuxPasswordPolicyFromDefs(tt.minLen, "90").Status)
		})
	}
}

func TestLinuxRemoteLoginStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state string
		want  Status
	}{
		{name: "active", state: "active", want: StatusFail},
		{name: "activating", state: "activating", want: StatusFail},
		{name: "inactive", state: "inactive", want: StatusPass},
		{name: "failed", state: "failed", want: StatusPass},
		{name: "deactivating", state: "deactivating", want: StatusPass},
		{name: "absent", state: "", want: StatusNotApplicable},
		{name: "reloading", state: "reloading", want: StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, linuxRemoteLoginStatus(tt.state))
		})
	}
}
