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
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	// Config bundles the deployment-level inputs the registry needs
	// to build typed providers: Probo's own AWS identity (used as
	// the AssumeRole source), the SSRF-protected HTTP client used
	// by GCP/Azure SDKs, and a logger.
	Config struct {
		BaseAWSConfig aws.Config
		HTTPClient    *http.Client
		Logger        *log.Logger
	}

	// Registry is the thread-safe factory for typed cloud-account
	// providers. It holds Probo's own AWS identity (used as the
	// STS AssumeRole source for AWS provider builds) and the
	// SSRF-protected HTTP client used by GCP/Azure SDKs. There is
	// no per-provider runtime state; building a Provider only
	// parses the supplied record's credentials envelope and wires
	// SDK clients pinned to that envelope.
	Registry struct {
		baseAWSConfig aws.Config
		httpClient    *http.Client
		logger        *log.Logger
	}
)

// NewRegistry returns a Registry built from cfg. The logger is
// optional; when nil a discard-output logger is used.
func NewRegistry(cfg Config) *Registry {
	logger := cfg.Logger
	if logger == nil {
		logger = log.NewLogger(log.WithOutput(io.Discard))
	}

	return &Registry{
		baseAWSConfig: cfg.BaseAWSConfig,
		httpClient:    cfg.HTTPClient,
		logger:        logger.Named("cloudaccount.registry"),
	}
}

// BuildAWSProvider returns a typed AWS provider pinned to the
// supplied record's AssumeRole credentials. Returns
// ErrCredentialsInvalid when the envelope is malformed or its kind
// is incompatible with provider=AWS.
func (r *Registry) BuildAWSProvider(rec CloudAccountRecord) (*AWSProvider, error) {
	if rec.Provider != coredata.CloudAccountProviderAWS {
		return nil, fmt.Errorf("cannot build aws provider: record provider is %q: %w", rec.Provider, ErrCredentialsInvalid)
	}

	creds, err := UnmarshalCredentials(coredata.CloudAccountProviderAWS, rec.DecryptedCredentials)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal aws credentials: %w", err)
	}

	awsCreds, ok := creds.(*AWSCredentials)
	if !ok {
		return nil, fmt.Errorf("cannot build aws provider: unexpected credentials type %T: %w", creds, ErrCredentialsInvalid)
	}

	return newAWSProvider(r.baseAWSConfig, rec, awsCreds), nil
}

// BuildGCPProvider returns a typed GCP provider pinned to the
// supplied record's service-account key. Returns
// ErrCredentialsInvalid when the envelope is malformed.
func (r *Registry) BuildGCPProvider(rec CloudAccountRecord) (*GCPProvider, error) {
	if rec.Provider != coredata.CloudAccountProviderGCP {
		return nil, fmt.Errorf("cannot build gcp provider: record provider is %q: %w", rec.Provider, ErrCredentialsInvalid)
	}

	creds, err := UnmarshalCredentials(coredata.CloudAccountProviderGCP, rec.DecryptedCredentials)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal gcp credentials: %w", err)
	}

	gcpCreds, ok := creds.(*GCPCredentials)
	if !ok {
		return nil, fmt.Errorf("cannot build gcp provider: unexpected credentials type %T: %w", creds, ErrCredentialsInvalid)
	}

	return newGCPProvider(r.httpClient, rec, gcpCreds), nil
}

// BuildAzureProvider returns a typed Azure provider pinned to the
// supplied record's client secret. Returns ErrCredentialsInvalid
// when the envelope is malformed.
func (r *Registry) BuildAzureProvider(rec CloudAccountRecord) (*AzureProvider, error) {
	if rec.Provider != coredata.CloudAccountProviderAzure {
		return nil, fmt.Errorf("cannot build azure provider: record provider is %q: %w", rec.Provider, ErrCredentialsInvalid)
	}

	creds, err := UnmarshalCredentials(coredata.CloudAccountProviderAzure, rec.DecryptedCredentials)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal azure credentials: %w", err)
	}

	azureCreds, ok := creds.(*AzureCredentials)
	if !ok {
		return nil, fmt.Errorf("cannot build azure provider: unexpected credentials type %T: %w", creds, ErrCredentialsInvalid)
	}

	return newAzureProvider(r.httpClient, rec, azureCreds), nil
}

// BuildProbeable is the only polymorphic entry point on Registry.
// It is used by the probe worker, which doesn't care about the
// concrete provider type. Internally it dispatches to the three
// typed builders.
func (r *Registry) BuildProbeable(rec CloudAccountRecord) (Probeable, error) {
	switch rec.Provider {
	case coredata.CloudAccountProviderAWS:
		return r.BuildAWSProvider(rec)
	case coredata.CloudAccountProviderGCP:
		return r.BuildGCPProvider(rec)
	case coredata.CloudAccountProviderAzure:
		return r.BuildAzureProvider(rec)
	default:
		return nil, fmt.Errorf("cannot build cloud account probeable: unsupported provider %q", rec.Provider)
	}
}
