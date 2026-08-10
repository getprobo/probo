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

package slackbot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionIDFor(t *testing.T) {
	t.Parallel()

	t.Run(
		"scopes channel threads by team channel and user",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				"T1:C1:111.000:U1",
				sessionIDFor("T1", "C1", ChannelTypeChannel, "111.000", "222.000", "U1"),
			)
			assert.Equal(
				t,
				"T1:C1:111.000:U2",
				sessionIDFor("T1", "C1", ChannelTypeChannel, "111.000", "222.000", "U2"),
			)
			assert.Equal(
				t,
				"T1:C2:111.000:U1",
				sessionIDFor("T1", "C2", ChannelTypeChannel, "111.000", "222.000", "U1"),
			)
			assert.Equal(
				t,
				"T2:C1:111.000:U1",
				sessionIDFor("T2", "C1", ChannelTypeChannel, "111.000", "222.000", "U1"),
			)
		},
	)

	t.Run(
		"falls back to message ts when thread is empty",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				"T1:C1:222.000:U1",
				sessionIDFor("T1", "C1", ChannelTypeChannel, "", "222.000", "U1"),
			)
		},
	)

	t.Run(
		"scopes im conversations by team and dm channel",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				"T1:D1",
				sessionIDFor("T1", "D1", ChannelTypeIM, "111.000", "222.000", "U1"),
			)
		},
	)

	t.Run(
		"returns empty when required fields are missing",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "", sessionIDFor("", "C1", ChannelTypeChannel, "111.000", "222.000", "U1"))
			assert.Equal(t, "", sessionIDFor("T1", "", ChannelTypeChannel, "111.000", "222.000", "U1"))
			assert.Equal(t, "", sessionIDFor("T1", "C1", ChannelTypeChannel, "", "", "U1"))
			assert.Equal(t, "", sessionIDFor("T1", "C1", ChannelTypeChannel, "111.000", "222.000", ""))
		},
	)
}
