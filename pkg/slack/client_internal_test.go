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
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// TestClientMethodURLMatchesCompiledSlackAPIBase pins the URLs this client
// emitted before the API base became a parameter. Threading the base must be a
// refactor for a deployment that configures no override: the compiled SLACK
// registration is the source of the base, and joining each method onto it has
// to reproduce the literals pkg/slack/client.go used to hold.
func TestClientMethodURLMatchesCompiledSlackAPIBase(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderSlack)
	require.True(t, ok)

	c := NewClient(reg.Endpoints.APIBase, log.NewLogger(log.WithName("test")))

	for method, want := range map[string]string{
		methodPostMessage:      "https://slack.com/api/chat.postMessage",
		methodUpdateMessage:    "https://slack.com/api/chat.update",
		methodConversationJoin: "https://slack.com/api/conversations.join",
	} {
		got, err := c.methodURL(method)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
}

// TestClientMethodURLFollowsOverriddenAPIBase is the regression test for the
// bug: the notification client used to dial slack.com no matter what, while
// the token it presented was minted by the OVERRIDDEN Endpoints.Token, so a
// sandbox deployment handed a sandbox-issued token to the real vendor.
func TestClientMethodURLFollowsOverriddenAPIBase(t *testing.T) {
	t.Parallel()

	const sandbox = "https://slack.sandbox.example.com/api"

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderSlack: {
			Auth:    "https://slack.sandbox.example.com/oauth/v2/authorize",
			Token:   "https://slack.sandbox.example.com/api/oauth.v2.access",
			Probe:   "https://slack.sandbox.example.com/api/users.list?limit=1",
			APIBase: sandbox,
		},
	}))
	require.NoError(t, err)

	reg, ok := r.Get(coredata.ConnectorProviderSlack)
	require.True(t, ok)

	c := NewClient(reg.Endpoints.APIBase, log.NewLogger(log.WithName("test")))

	got, err := c.methodURL(methodPostMessage)
	require.NoError(t, err)
	assert.Equal(t, sandbox+"/chat.postMessage", got)
}

// TestValidateSlackResponseURLIgnoresAPIBase keeps the webhook host check what
// it is: a guard on the response_url an INBOUND interaction payload carries,
// not an outbound endpoint the deployment chooses. Slack serves response URLs
// from hooks.slack.com, a host no Endpoints field names, so an override must
// not widen it — that would turn a configurable base into an SSRF sink.
func TestValidateSlackResponseURLIgnoresAPIBase(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateSlackResponseURL("https://hooks.slack.com/actions/T0/1/abc"))

	for _, responseURL := range []string{
		"https://slack.sandbox.example.com/actions/T0/1/abc",
		"https://attacker.example.com/actions/T0/1/abc",
		"http://hooks.slack.com/actions/T0/1/abc",
	} {
		assert.Error(t, validateSlackResponseURL(responseURL), "response URL %q must be refused", responseURL)
	}
}
