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

	"go.probo.inc/probo/pkg/coredata"
)

// CredentialsEnvelopeVersion is the persisted JSON envelope version.
// Bump only when the on-disk shape changes incompatibly; a rotation
// in the at-rest cipher key version (the 0x01 prefix byte) does NOT
// change this constant.
const CredentialsEnvelopeVersion = 1

type (
	// CloudAccountRecord is the thin value type pkg/cloudaccount
	// operates on. The service layer (pkg/probo) maps a
	// *coredata.CloudAccount to this record before calling the
	// registry, keeping pkg/cloudaccount free of any data-layer
	// entity dependency.
	CloudAccountRecord struct {
		ID                   string
		Provider             coredata.CloudAccountProvider
		Kind                 coredata.CloudAccountCredentialKind
		ScopeKind            coredata.CloudAccountScopeKind
		ScopeIdentifier      string
		ExternalID           string
		DecryptedCredentials []byte
	}

	// Credentials is the polymorphic per-(provider, kind) payload
	// carried inside CloudAccountRecord.DecryptedCredentials. The
	// kind lives inside the JSON envelope (see UnmarshalCredentials)
	// -- it is NOT passed as a separate parameter to the unmarshal
	// entry point.
	Credentials interface {
		Provider() coredata.CloudAccountProvider
		Kind() coredata.CloudAccountCredentialKind
		json.Marshaler
		json.Unmarshaler
	}

	// Probeable is the only behaviour the worker depends on. Each
	// typed *AWSProvider / *GCPProvider / *AzureProvider satisfies
	// it via its own Probe method. Drivers depend on narrower seam
	// interfaces declared in driver files.
	Probeable interface {
		Probe(ctx context.Context) error
	}

	// credentialsEnvelope is the on-disk JSON shape every typed
	// *Credentials marshals to. v is the envelope version; kind is
	// the source-of-truth credential kind; payload is the
	// per-(provider, kind) inner value.
	credentialsEnvelope struct {
		V       int                                 `json:"v"`
		Kind    coredata.CloudAccountCredentialKind `json:"kind"`
		Payload json.RawMessage                     `json:"payload"`
	}
)

// providerForKind names the only provider compatible with a given
// credential kind. Used by UnmarshalCredentials to defeat
// provider/kind mismatch envelopes (e.g. an Azure payload submitted
// against an AWS row).
func providerForKind(k coredata.CloudAccountCredentialKind) (coredata.CloudAccountProvider, bool) {
	switch k {
	case coredata.CloudAccountCredentialKindAWSAssumeRole:
		return coredata.CloudAccountProviderAWS, true
	case coredata.CloudAccountCredentialKindGCPServiceAccountKey,
		coredata.CloudAccountCredentialKindGCPWorkloadIdentityFederation:
		return coredata.CloudAccountProviderGCP, true
	case coredata.CloudAccountCredentialKindAzureClientSecret,
		coredata.CloudAccountCredentialKindAzureFederatedCredential:
		return coredata.CloudAccountProviderAzure, true
	default:
		return "", false
	}
}

// MarshalEnvelope wraps a credentials payload in the versioned
// {"v":1,"kind":"...","payload":{...}} JSON envelope every typed
// *Credentials uses for on-disk persistence.
func MarshalEnvelope(kind coredata.CloudAccountCredentialKind, payload any) ([]byte, error) {
	inner, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal cloud account credentials payload: %w", err)
	}

	env := credentialsEnvelope{
		V:       CredentialsEnvelopeVersion,
		Kind:    kind,
		Payload: inner,
	}

	out, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal cloud account credentials envelope: %w", err)
	}

	return out, nil
}

// UnmarshalCredentials parses the versioned envelope and returns
// the typed credentials. The envelope's "kind" is the source of
// truth; the caller does NOT pass kind separately. The expected
// provider parameter is cross-checked against providerForKind to
// reject envelopes whose kind belongs to a different provider
// (e.g. an Azure payload submitted under provider=AWS) with
// ErrCredentialsInvalid.
func UnmarshalCredentials(provider coredata.CloudAccountProvider, payload []byte) (Credentials, error) {
	var env credentialsEnvelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return nil, fmt.Errorf("cannot unmarshal cloud account credentials envelope: %w", ErrCredentialsInvalid)
	}

	if env.V != CredentialsEnvelopeVersion {
		return nil, fmt.Errorf("cannot unmarshal cloud account credentials: unsupported envelope version %d: %w", env.V, ErrCredentialsInvalid)
	}

	expected, ok := providerForKind(env.Kind)
	if !ok {
		return nil, fmt.Errorf("cannot unmarshal cloud account credentials: unknown kind %q: %w", env.Kind, ErrCredentialsInvalid)
	}

	if expected != provider {
		return nil, fmt.Errorf("cannot unmarshal cloud account credentials: kind %q is incompatible with provider %q: %w", env.Kind, provider, ErrCredentialsInvalid)
	}

	switch env.Kind {
	case coredata.CloudAccountCredentialKindAWSAssumeRole:
		var c AWSCredentials
		if err := json.Unmarshal(payload, &c); err != nil {
			return nil, fmt.Errorf("cannot unmarshal aws credentials: %w", err)
		}

		return &c, nil
	case coredata.CloudAccountCredentialKindGCPServiceAccountKey:
		var c GCPCredentials
		if err := json.Unmarshal(payload, &c); err != nil {
			return nil, fmt.Errorf("cannot unmarshal gcp credentials: %w", err)
		}

		return &c, nil
	case coredata.CloudAccountCredentialKindAzureClientSecret:
		var c AzureCredentials
		if err := json.Unmarshal(payload, &c); err != nil {
			return nil, fmt.Errorf("cannot unmarshal azure credentials: %w", err)
		}

		return &c, nil
	default:
		return nil, fmt.Errorf("cannot unmarshal cloud account credentials: unsupported kind %q: %w", env.Kind, ErrCredentialsInvalid)
	}
}
