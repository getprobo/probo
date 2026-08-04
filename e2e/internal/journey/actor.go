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

package journey

import (
	"fmt"

	"go.probo.inc/probo/e2e/internal/testutil"
)

// Actor is a named participant with a real authenticated or unauthenticated
// end-to-end client.
type Actor struct {
	name   string
	world  *World
	client *testutil.Client
}

// NewActor signs up a real user and creates an organization through the
// existing end-to-end client flow.
func (w *World) NewActor(name string, role testutil.TestRole) *Actor {
	w.t.Helper()

	var client *testutil.Client
	w.Step(name, fmt.Sprintf("signs up as %s and creates an organization", role), func() error {
		client = testutil.NewClient(w.t, role)
		return nil
	})

	return w.registerClient(name, client)
}

// NewMemberActor creates, invites, activates, and signs in a real member of the
// owner's organization through the existing end-to-end client flow.
func (w *World) NewMemberActor(
	name string,
	role testutil.TestRole,
	owner *Actor,
) *Actor {
	w.t.Helper()

	var client *testutil.Client
	w.Step(name, fmt.Sprintf("joins %s's organization as %s", owner.name, role), func() error {
		client = testutil.NewClientInOrg(w.t, role, owner.client)
		return nil
	})

	return w.registerClient(name, client)
}

// NewUnauthenticatedActor creates a real cookie-backed client without signing
// in. It is useful for onboarding and authorization journeys.
func (w *World) NewUnauthenticatedActor(name string) *Actor {
	w.t.Helper()

	client := testutil.NewUnauthenticatedClient(w.t)

	return w.registerClient(name, client)
}

// Name returns the human-readable actor name used in diagnostics.
func (a *Actor) Name() string {
	return a.name
}

// Client returns the underlying black-box client for factories and specialized
// protocol helpers.
func (a *Actor) Client() *testutil.Client {
	return a.client
}

// Step executes one action attributed to this actor.
func (a *Actor) Step(name string, fn func() error) {
	a.world.t.Helper()
	a.world.Step(a.name, name, fn)
}

func (w *World) registerClient(name string, client *testutil.Client) *Actor {
	actor := &Actor{
		name:   name,
		world:  w,
		client: client,
	}

	record := actorRecord{
		Name: name,
		Role: string(client.GetRole()),
	}
	if client.GetRole() != "" {
		record.UserID = client.GetUserID().String()
		record.OrganizationID = client.GetOrganizationID().String()
	}

	w.registerActor(record)

	return actor
}
