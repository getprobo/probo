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

package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvelopeInstallationTeamID(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"T_CONTEXT",
		(Envelope{
			TeamID:        "T_ACTOR",
			ContextTeamID: "T_CONTEXT",
			Authorizations: []Authorization{{
				TeamID: "T_AUTHORIZATION",
			}},
		}).InstallationTeamID(),
	)
	assert.Equal(
		t,
		"T_AUTHORIZATION",
		(Envelope{
			TeamID: "T_ACTOR",
			Authorizations: []Authorization{{
				TeamID: "T_AUTHORIZATION",
			}},
		}).InstallationTeamID(),
	)
}

func TestEventActorTeamID(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"T_EXTERNAL",
		(EventBody{UserTeam: "T_EXTERNAL"}).ActorTeamID("T_INSTALL"),
	)
	assert.Equal(
		t,
		"T_INSTALL",
		(EventBody{}).ActorTeamID("T_INSTALL"),
	)
}
