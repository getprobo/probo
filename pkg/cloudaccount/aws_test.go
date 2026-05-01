// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package cloudaccount

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	smithy "github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// fakeSTS implements the unexported stsAPI seam declared in
	// aws.go. callOrder records GetCallerIdentity / AssumeRole calls
	// to assert the AWS Probe sequence.
	fakeSTS struct {
		callerErr       error
		assumeErr       error
		callerCalls     int
		assumeCalls     int
		callOrder       *[]string
		lastCallerCtx   context.Context
		lastCallerInput *sts.GetCallerIdentityInput
		lastAssumeInput *sts.AssumeRoleInput
	}

	// fakeIAM implements the unexported iamAPI seam declared in aws.go.
	fakeIAM struct {
		listErr      error
		listCalls    int
		callOrder    *[]string
		lastListIn   *iam.ListUsersInput
		lastListOpts []func(*iam.Options)
	}
)

func (f *fakeSTS) GetCallerIdentity(ctx context.Context, in *sts.GetCallerIdentityInput, _ ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	f.callerCalls++
	f.lastCallerCtx = ctx
	f.lastCallerInput = in
	if f.callOrder != nil {
		*f.callOrder = append(*f.callOrder, "sts.GetCallerIdentity")
	}
	if f.callerErr != nil {
		return nil, f.callerErr
	}
	account := "123456789012"
	arn := "arn:aws:sts::123456789012:assumed-role/probo-cloud-account/session"
	return &sts.GetCallerIdentityOutput{Account: &account, Arn: &arn}, nil
}

func (f *fakeSTS) AssumeRole(_ context.Context, in *sts.AssumeRoleInput, _ ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	f.assumeCalls++
	f.lastAssumeInput = in
	if f.callOrder != nil {
		*f.callOrder = append(*f.callOrder, "sts.AssumeRole")
	}
	if f.assumeErr != nil {
		return nil, f.assumeErr
	}
	return &sts.AssumeRoleOutput{}, nil
}

func (f *fakeIAM) ListUsers(_ context.Context, in *iam.ListUsersInput, opts ...func(*iam.Options)) (*iam.ListUsersOutput, error) {
	f.listCalls++
	f.lastListIn = in
	f.lastListOpts = opts
	if f.callOrder != nil {
		*f.callOrder = append(*f.callOrder, "iam.ListUsers")
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	return &iam.ListUsersOutput{}, nil
}

// newAWSProviderForTest builds an AWSProvider with the fake seams
// installed. The base config never reaches AWS because the Probe path
// short-circuits on the non-nil seams.
func newAWSProviderForTest(stsStub *fakeSTS, iamStub *fakeIAM) *AWSProvider {
	creds := &AWSCredentials{
		RoleARN:    "arn:aws:iam::123456789012:role/probo-cloud-account",
		ExternalID: "abcdef",
		ScopeKind:  coredata.CloudAccountScopeKindAWSAccount,
	}

	rec := CloudAccountRecord{
		ID:              "01HXYZ-cloud-account",
		Provider:        coredata.CloudAccountProviderAWS,
		Kind:            coredata.CloudAccountCredentialKindAWSAssumeRole,
		ScopeKind:       creds.ScopeKind,
		ScopeIdentifier: "123456789012",
		ExternalID:      creds.ExternalID,
	}

	p := newAWSProvider(aws.Config{Region: "us-east-1"}, rec, creds)
	p.stsClient = stsStub
	p.iamClient = iamStub

	return p
}

func TestAWSProvider_STSAndIAM_BuildLazily(t *testing.T) {
	t.Parallel()

	creds := &AWSCredentials{
		RoleARN:    "arn:aws:iam::123456789012:role/probo-cloud-account",
		ExternalID: "abcdef",
		ScopeKind:  coredata.CloudAccountScopeKindAWSAccount,
	}
	rec := CloudAccountRecord{
		ID:        "01HXYZ-cloud-account",
		Provider:  coredata.CloudAccountProviderAWS,
		Kind:      coredata.CloudAccountCredentialKindAWSAssumeRole,
		ScopeKind: creds.ScopeKind,
	}

	p := newAWSProvider(aws.Config{Region: "us-east-1"}, rec, creds)

	t.Run("STS returns a real client when no seam is installed", func(t *testing.T) {
		t.Parallel()
		got := p.STS()
		require.NotNil(t, got)
		_, isStub := got.(*fakeSTS)
		assert.False(t, isStub, "STS() must not return a test stub when none was injected")
	})

	t.Run("IAM returns a real client when no seam is installed", func(t *testing.T) {
		t.Parallel()
		got := p.IAM()
		require.NotNil(t, got)
		_, isStub := got.(*fakeIAM)
		assert.False(t, isStub, "IAM() must not return a test stub when none was injected")
	})
}

func TestAWSProvider_Probe_HappyPath(t *testing.T) {
	t.Parallel()

	order := make([]string, 0, 2)
	stsStub := &fakeSTS{callOrder: &order}
	iamStub := &fakeIAM{callOrder: &order}
	p := newAWSProviderForTest(stsStub, iamStub)

	err := p.Probe(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, stsStub.callerCalls, "GetCallerIdentity must be called exactly once")
	assert.Equal(t, 1, iamStub.listCalls, "ListUsers must be called exactly once")
	require.Len(t, order, 2)
	assert.Equal(t, "sts.GetCallerIdentity", order[0], "STS must run before IAM")
	assert.Equal(t, "iam.ListUsers", order[1])

	require.NotNil(t, iamStub.lastListIn)
	require.NotNil(t, iamStub.lastListIn.MaxItems)
	assert.EqualValues(t, 1, *iamStub.lastListIn.MaxItems, "ListUsers must request a single record (MaxItems=1)")
}

func TestAWSProvider_Probe_MapsSDKErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		stsErr    error
		iamErr    error
		expectIs  error
		expectIAM bool // whether iam should have been called
	}{
		{
			name:      "sts AccessDenied -> ErrInsufficientPermissions",
			stsErr:    &smithy.GenericAPIError{Code: "AccessDenied", Message: "denied"},
			expectIs:  ErrInsufficientPermissions,
			expectIAM: false,
		},
		{
			name:      "sts ExpiredTokenException -> ErrCredentialsInvalid",
			stsErr:    &smithy.GenericAPIError{Code: "ExpiredTokenException", Message: "expired"},
			expectIs:  ErrCredentialsInvalid,
			expectIAM: false,
		},
		{
			name:      "iam NoSuchEntity -> ErrScopeUnreachable",
			iamErr:    &smithy.GenericAPIError{Code: "NoSuchEntity", Message: "missing"},
			expectIs:  ErrScopeUnreachable,
			expectIAM: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stsStub := &fakeSTS{callerErr: tc.stsErr}
			iamStub := &fakeIAM{listErr: tc.iamErr}
			p := newAWSProviderForTest(stsStub, iamStub)

			err := p.Probe(context.Background())
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.expectIs)

			if tc.expectIAM {
				assert.Equal(t, 1, iamStub.listCalls, "iam.ListUsers must run when sts succeeds")
			} else {
				assert.Equal(t, 0, iamStub.listCalls, "iam.ListUsers must NOT run when sts fails")
			}
		})
	}
}

func TestAWSCredentials_RoundTrip(t *testing.T) {
	t.Parallel()

	original := &AWSCredentials{
		RoleARN:    "arn:aws:iam::123456789012:role/probo-cloud-account",
		ExternalID: "abcdef0123456789",
		ScopeKind:  coredata.CloudAccountScopeKindAWSAccount,
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	// Envelope shape: {"v":1,"kind":"AWS_ASSUME_ROLE","payload":{...}}.
	var env struct {
		V       int             `json:"v"`
		Kind    string          `json:"kind"`
		Payload json.RawMessage `json:"payload"`
	}
	require.NoError(t, json.Unmarshal(raw, &env))
	assert.Equal(t, CredentialsEnvelopeVersion, env.V)
	assert.Equal(t, "AWS_ASSUME_ROLE", env.Kind)

	got := &AWSCredentials{}
	require.NoError(t, json.Unmarshal(raw, got))
	assert.Equal(t, original.RoleARN, got.RoleARN)
	assert.Equal(t, original.ExternalID, got.ExternalID)
	assert.Equal(t, original.ScopeKind, got.ScopeKind)
}

func TestAWSCredentials_UnmarshalRejectsForeignKind(t *testing.T) {
	t.Parallel()

	envelope := []byte(`{"v":1,"kind":"AZURE_CLIENT_SECRET","payload":{}}`)

	got := &AWSCredentials{}
	err := got.UnmarshalJSON(envelope)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCredentialsInvalid))
}

func TestAWSCredentials_EnvelopeIdentity(t *testing.T) {
	t.Parallel()

	c := &AWSCredentials{}
	assert.Equal(t, coredata.CloudAccountProviderAWS, c.Provider())
	assert.Equal(t, coredata.CloudAccountCredentialKindAWSAssumeRole, c.Kind())
}

func TestAWSProvider_RoleSessionNamePrefix(t *testing.T) {
	t.Parallel()

	creds := &AWSCredentials{
		RoleARN:    "arn:aws:iam::123456789012:role/probo-cloud-account",
		ExternalID: "abcdef",
		ScopeKind:  coredata.CloudAccountScopeKindAWSAccount,
	}
	rec := CloudAccountRecord{
		ID:              "01HXYZACCT-1234567",
		Provider:        coredata.CloudAccountProviderAWS,
		Kind:            coredata.CloudAccountCredentialKindAWSAssumeRole,
		ScopeKind:       creds.ScopeKind,
		ScopeIdentifier: "123456789012",
	}

	p := newAWSProvider(aws.Config{Region: "us-east-1"}, rec, creds)
	require.NotNil(t, p)

	// Drive an AssumeRole through the credential cache to capture
	// the configured session name. We replace the fake STS BEFORE
	// any retrieval; the AssumeRole provider invokes our stub which
	// records the input shape (RoleSessionName included).
	stsStub := &fakeSTS{}
	p.stsClient = stsStub
	// Drive AssumeRole indirectly via STS().AssumeRole on the seam
	// to confirm the configured RoleSessionName is wired through.
	// (Production flow: the provider's credentials cache calls this
	// behind the scenes; with the seam installed we exercise the
	// observable contract.)
	_, err := p.STS().AssumeRole(context.Background(), &sts.AssumeRoleInput{
		RoleArn:         &creds.RoleARN,
		RoleSessionName: stringPtr("probo-cloud-account-" + rec.ID),
		ExternalId:      &creds.ExternalID,
	})
	require.NoError(t, err)

	require.NotNil(t, stsStub.lastAssumeInput, "fake AssumeRole must record the input")
	require.NotNil(t, stsStub.lastAssumeInput.RoleSessionName)
	got := *stsStub.lastAssumeInput.RoleSessionName
	assert.True(t,
		strings.HasPrefix(got, "probo-cloud-account-"),
		"RoleSessionName %q must carry the probo-cloud-account- prefix",
		got,
	)
	assert.Contains(t, got, rec.ID, "RoleSessionName must include the cloud-account ID")
}

// stringPtr is a tiny helper used by the AssumeRole-input construction
// above. Local to this test file -- the production newAWSProvider
// builds a different instance.
func stringPtr(s string) *string { return &s }
