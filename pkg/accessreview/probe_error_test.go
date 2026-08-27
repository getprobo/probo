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

package accessreview_test

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/aws/smithy-go"
	"golang.org/x/oauth2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestProbeErrorCarriesProvider(t *testing.T) {
	t.Parallel()

	cause := errors.New("boom")
	err := accessreview.NewProbeError(coredata.ConnectorProviderLangfuse, cause)

	probeErr, ok := errors.AsType[*accessreview.ProbeError](err)
	require.True(t, ok)
	assert.Equal(t, coredata.ConnectorProviderLangfuse, probeErr.Provider)
	// Unwrap keeps errors.Is working through the wrapper, which the resolver
	// relies on to still recognise a missing connector.
	assert.ErrorIs(t, err, cause)
}

func TestProbeFailureCodeIsSafeToLog(t *testing.T) {
	t.Parallel()

	// A provider that refuses the credential is the one case worth reporting
	// precisely, because the status tells an operator whether to reconnect.
	rejected := accessreview.NewProbeError(
		coredata.ConnectorProviderLangfuse,
		&provider.CredentialRejectedError{StatusCode: 403},
	)
	assert.Equal(t, "credential_rejected_403", accessreview.ProbeFailureCode(rejected))

	// A transport error embeds the customer's self-hosted host, so only the
	// classification survives.
	transport := accessreview.NewProbeError(
		coredata.ConnectorProviderLangfuse,
		&url.Error{
			Op:  "Get",
			URL: "https://langfuse.internal.customer.example/api/public/organizations/memberships",
			Err: errors.New("dial tcp: connection refused"),
		},
	)
	code := accessreview.ProbeFailureCode(transport)
	assert.Equal(t, "transport_error", code)
	assert.NotContains(t, code, "customer.example")

	// Anything unrecognised degrades to its type, which names the failure
	// without quoting provider-controlled text.
	opaque := accessreview.NewProbeError(
		coredata.ConnectorProviderLangfuse,
		fmt.Errorf("cannot refresh token: %w", errors.New(`oauth2: "invalid_grant" "token revoked for user ada@example.com"`)),
	)
	opaqueCode := accessreview.ProbeFailureCode(opaque)
	assert.NotContains(t, opaqueCode, "ada@example.com")
	assert.NotContains(t, opaqueCode, "invalid_grant")

	// An AWS error code is a fixed identifier, so it is reported; the message
	// around it, which can name a role ARN, is not.
	awsCode := accessreview.ProbeFailureCode(
		fmt.Errorf("cannot reach aws account: %w", &smithy.GenericAPIError{
			Code:    "AccessDenied",
			Message: "User: arn:aws:sts::123456789012:assumed-role/probo is not authorized",
		}),
	)
	assert.Equal(t, "aws_AccessDenied", awsCode)
	assert.NotContains(t, awsCode, "123456789012")
}

func TestProbeFailureCodeSurvivesTypedNil(t *testing.T) {
	t.Parallel()

	// ProbeFailureCode runs while logging a failure, so a malformed error must
	// not turn a degraded connector into a panicking resolver.
	var rejected *provider.CredentialRejectedError

	assert.NotPanics(t, func() {
		accessreview.ProbeFailureCode(accessreview.NewProbeError(coredata.ConnectorProviderLangfuse, rejected))
	})
}

func TestIsProviderVerdict(t *testing.T) {
	t.Parallel()

	// The provider answered.
	assert.True(t, accessreview.IsProviderVerdict(&provider.CredentialRejectedError{StatusCode: 401}))
	assert.True(t, accessreview.IsProviderVerdict(&url.Error{Op: "Get", URL: "https://x.example", Err: errors.New("refused")}))
	assert.True(t, accessreview.IsProviderVerdict(&oauth2.RetrieveError{ErrorCode: "invalid_grant"}))
	// A workload identity connector is answered by STS through the SDK, so an
	// AWS API error is the provider's verdict, not a Probo failure.
	assert.True(t, accessreview.IsProviderVerdict(fmt.Errorf("cannot reach aws account: %w", &smithy.GenericAPIError{Code: "AccessDenied", Message: "not authorized"})))

	// Probo never got as far as asking. Defaulting these to "ours" keeps a
	// settings decode or a request we could not build in the error budget,
	// with its message, instead of being blamed on the customer's credential.
	assert.False(t, accessreview.IsProviderVerdict(errors.New("cannot read crisp connector settings: unexpected end of JSON input")))
	assert.False(t, accessreview.IsProviderVerdict(fmt.Errorf("cannot build probe URL: %w", errors.New("missing crisp website_id"))))
	assert.False(t, accessreview.IsProviderVerdict(errors.New("cannot persist refreshed token: connection reset")))
}
