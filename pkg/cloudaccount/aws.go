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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	// AWSCredentials carry the inputs Probo needs to AssumeRole into
	// the customer account. v1 deliberately omits a Regions field --
	// IAM is global, the v1 access-review fetch only touches iam:*
	// and sts:*, and a regional list field would be unused state.
	// Reintroduce alongside the first regional CSPM module (e.g.
	// S3 public-bucket scan) so the schema change is paired with an
	// actual consumer.
	AWSCredentials struct {
		RoleARN    string                         `json:"role_arn"`
		ExternalID string                         `json:"external_id"`
		ScopeKind  coredata.CloudAccountScopeKind `json:"scope_kind"`
	}

	// stsAPI is the narrow seam AWSProvider depends on for STS
	// calls. The real *sts.Client satisfies it implicitly; tests
	// inject a stub.
	stsAPI interface {
		GetCallerIdentity(ctx context.Context, in *sts.GetCallerIdentityInput, opts ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
		AssumeRole(ctx context.Context, in *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
	}

	// iamAPI is the narrow seam AWSProvider depends on for IAM
	// calls. The real *iam.Client satisfies it implicitly; tests
	// inject a stub.
	iamAPI interface {
		ListUsers(ctx context.Context, in *iam.ListUsersInput, opts ...func(*iam.Options)) (*iam.ListUsersOutput, error)
	}

	// AWSProvider builds typed AWS SDK clients pinned to a single
	// CloudAccountRecord's AssumeRole credentials.
	AWSProvider struct {
		baseConfig  aws.Config
		assumedCfg  aws.Config
		record      CloudAccountRecord
		credentials *AWSCredentials

		// Test seams. When non-nil, these override the lazily-built
		// real SDK clients. Production builds leave them nil.
		stsClient stsAPI
		iamClient iamAPI
	}
)

// Compile-time interface assertions.
var (
	_ Credentials = (*AWSCredentials)(nil)
	_ Probeable   = (*AWSProvider)(nil)
)

func (c *AWSCredentials) Provider() coredata.CloudAccountProvider {
	return coredata.CloudAccountProviderAWS
}

func (c *AWSCredentials) Kind() coredata.CloudAccountCredentialKind {
	return coredata.CloudAccountCredentialKindAWSAssumeRole
}

// MarshalJSON wraps the AWSCredentials payload in the versioned
// envelope. The "v" and "kind" envelope fields are stamped here so
// callers cannot accidentally persist a kind that disagrees with
// the typed value.
func (c *AWSCredentials) MarshalJSON() ([]byte, error) {
	return MarshalEnvelope(c.Kind(), struct {
		RoleARN    string                         `json:"role_arn"`
		ExternalID string                         `json:"external_id"`
		ScopeKind  coredata.CloudAccountScopeKind `json:"scope_kind"`
	}{
		RoleARN:    c.RoleARN,
		ExternalID: c.ExternalID,
		ScopeKind:  c.ScopeKind,
	})
}

// UnmarshalJSON accepts either the bare payload (used internally
// by UnmarshalCredentials after envelope stripping) or the full
// envelope shape. It rejects non-AWS_ASSUME_ROLE envelopes with
// ErrCredentialsInvalid.
func (c *AWSCredentials) UnmarshalJSON(data []byte) error {
	var env credentialsEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("cannot unmarshal aws credentials envelope: %w", err)
	}

	if env.V == 0 && env.Kind == "" {
		// Bare payload (no envelope wrapping).
		var payload struct {
			RoleARN    string                         `json:"role_arn"`
			ExternalID string                         `json:"external_id"`
			ScopeKind  coredata.CloudAccountScopeKind `json:"scope_kind"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("cannot unmarshal aws credentials payload: %w", err)
		}

		c.RoleARN = payload.RoleARN
		c.ExternalID = payload.ExternalID
		c.ScopeKind = payload.ScopeKind

		return nil
	}

	if env.Kind != coredata.CloudAccountCredentialKindAWSAssumeRole {
		return fmt.Errorf("cannot unmarshal aws credentials: kind %q: %w", env.Kind, ErrCredentialsInvalid)
	}

	var payload struct {
		RoleARN    string                         `json:"role_arn"`
		ExternalID string                         `json:"external_id"`
		ScopeKind  coredata.CloudAccountScopeKind `json:"scope_kind"`
	}
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		return fmt.Errorf("cannot unmarshal aws credentials payload: %w", err)
	}

	c.RoleARN = payload.RoleARN
	c.ExternalID = payload.ExternalID
	c.ScopeKind = payload.ScopeKind

	return nil
}

// newAWSProvider is the package-internal constructor invoked by
// Registry.BuildAWSProvider. It builds an aws.Config whose
// credential provider is an stscreds.AssumeRoleProvider over the
// supplied baseConfig (Probo's STS identity), pinned to the record's
// RoleARN/ExternalID and a deterministic RoleSessionName.
func newAWSProvider(baseConfig aws.Config, rec CloudAccountRecord, creds *AWSCredentials) *AWSProvider {
	stsClient := sts.NewFromConfig(baseConfig)

	provider := stscreds.NewAssumeRoleProvider(stsClient, creds.RoleARN, func(o *stscreds.AssumeRoleOptions) {
		o.ExternalID = new(creds.ExternalID)
		o.RoleSessionName = fmt.Sprintf("probo-cloud-account-%s", rec.ID)
	})

	assumed := baseConfig.Copy()
	assumed.Credentials = aws.NewCredentialsCache(provider)

	return &AWSProvider{
		baseConfig:  baseConfig,
		assumedCfg:  assumed,
		record:      rec,
		credentials: creds,
	}
}

// STS returns an *sts.Client (or test stub) bound to the record's
// AssumeRole credentials.
func (p *AWSProvider) STS() stsAPI {
	if p.stsClient != nil {
		return p.stsClient
	}

	return sts.NewFromConfig(p.assumedCfg)
}

// IAM returns an *iam.Client (or test stub) bound to the record's
// AssumeRole credentials.
func (p *AWSProvider) IAM() iamAPI {
	if p.iamClient != nil {
		return p.iamClient
	}

	return iam.NewFromConfig(p.assumedCfg)
}

// Probe verifies the AssumeRole + IAM read paths against the
// customer account by issuing sts:GetCallerIdentity followed by
// iam:ListUsers (MaxItems=1). Errors are mapped to typed package
// sentinels via MapSDKError.
func (p *AWSProvider) Probe(ctx context.Context) error {
	if _, err := p.STS().GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}); err != nil {
		return MapSDKError(fmt.Errorf("cannot probe aws caller identity: %w", err))
	}

	maxItems := int32(1)
	if _, err := p.IAM().ListUsers(ctx, &iam.ListUsersInput{MaxItems: &maxItems}); err != nil {
		return MapSDKError(fmt.Errorf("cannot probe aws iam users: %w", err))
	}

	return nil
}

// Record returns the underlying CloudAccountRecord. Useful for
// drivers that need to discriminate on ScopeKind without
// re-threading the value through their constructor.
func (p *AWSProvider) Record() CloudAccountRecord {
	return p.record
}
