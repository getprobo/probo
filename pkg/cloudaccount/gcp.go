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
	"net/http"

	"golang.org/x/oauth2/google"
	cloudasset "google.golang.org/api/cloudasset/v1"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	gcpCloudPlatformReadOnlyScope = "https://www.googleapis.com/auth/cloud-platform.read-only"
)

type (
	// GCPCredentials carry the v1 service-account JSON key bytes.
	// The v2 placeholder for workload-identity federation will live
	// in the same struct (or a sibling) gated on Kind.
	GCPCredentials struct {
		ServiceAccountJSON []byte                         `json:"service_account_json"`
		ScopeKind          coredata.CloudAccountScopeKind `json:"scope_kind"`
		ProjectID          string                         `json:"project_id,omitempty"`
		OrganizationID     string                         `json:"organization_id,omitempty"`
	}

	// crmProjectsAPI is the narrow seam GCPProvider depends on for
	// project IAM-policy reads. The real
	// *cloudresourcemanager.ProjectsService satisfies it via its
	// Get(...).Do() chain; tests inject a stub.
	crmProjectsAPI interface {
		Get(ctx context.Context, projectID string) (*cloudresourcemanager.Project, error)
	}

	// crmOrganizationsAPI is the narrow seam GCPProvider depends on
	// for org reads.
	crmOrganizationsAPI interface {
		Get(ctx context.Context, name string) (*cloudresourcemanager.Organization, error)
	}

	// GCPProvider builds typed GCP SDK clients pinned to a single
	// CloudAccountRecord's service-account key.
	GCPProvider struct {
		httpClient  *http.Client
		record      CloudAccountRecord
		credentials *GCPCredentials

		// Test seams. Production builds leave these nil and the
		// real SDK clients are built lazily.
		crmProjects      crmProjectsAPI
		crmOrganizations crmOrganizationsAPI
	}
)

// Compile-time interface assertions.
var (
	_ Credentials = (*GCPCredentials)(nil)
	_ Probeable   = (*GCPProvider)(nil)
)

func (c *GCPCredentials) Provider() coredata.CloudAccountProvider {
	return coredata.CloudAccountProviderGCP
}

func (c *GCPCredentials) Kind() coredata.CloudAccountCredentialKind {
	return coredata.CloudAccountCredentialKindGCPServiceAccountKey
}

// MarshalJSON wraps the GCP credentials payload in the versioned
// envelope.
func (c *GCPCredentials) MarshalJSON() ([]byte, error) {
	return MarshalEnvelope(c.Kind(), struct {
		ServiceAccountJSON []byte                         `json:"service_account_json"`
		ScopeKind          coredata.CloudAccountScopeKind `json:"scope_kind"`
		ProjectID          string                         `json:"project_id,omitempty"`
		OrganizationID     string                         `json:"organization_id,omitempty"`
	}{
		ServiceAccountJSON: c.ServiceAccountJSON,
		ScopeKind:          c.ScopeKind,
		ProjectID:          c.ProjectID,
		OrganizationID:     c.OrganizationID,
	})
}

// UnmarshalJSON accepts either the bare payload or the full
// envelope shape; rejects non-GCP-compatible kinds with
// ErrCredentialsInvalid.
func (c *GCPCredentials) UnmarshalJSON(data []byte) error {
	var env credentialsEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("cannot unmarshal gcp credentials envelope: %w", err)
	}

	payloadBytes := data
	if env.V != 0 || env.Kind != "" {
		if env.Kind != coredata.CloudAccountCredentialKindGCPServiceAccountKey {
			return fmt.Errorf("cannot unmarshal gcp credentials: kind %q: %w", env.Kind, ErrCredentialsInvalid)
		}
		payloadBytes = env.Payload
	}

	var payload struct {
		ServiceAccountJSON []byte                         `json:"service_account_json"`
		ScopeKind          coredata.CloudAccountScopeKind `json:"scope_kind"`
		ProjectID          string                         `json:"project_id,omitempty"`
		OrganizationID     string                         `json:"organization_id,omitempty"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("cannot unmarshal gcp credentials payload: %w", err)
	}

	c.ServiceAccountJSON = payload.ServiceAccountJSON
	c.ScopeKind = payload.ScopeKind
	c.ProjectID = payload.ProjectID
	c.OrganizationID = payload.OrganizationID

	return nil
}

// newGCPProvider is the package-internal constructor invoked by
// Registry.BuildGCPProvider.
func newGCPProvider(httpClient *http.Client, rec CloudAccountRecord, creds *GCPCredentials) *GCPProvider {
	return &GCPProvider{
		httpClient:  httpClient,
		record:      rec,
		credentials: creds,
	}
}

// GoogleCredentials returns a *google.Credentials parsed from the
// record's service-account JSON key. Callers that need a low-level
// token source use this; most callers prefer the typed helpers
// (CRM, CloudAsset).
func (p *GCPProvider) GoogleCredentials(ctx context.Context) (*google.Credentials, error) {
	creds, err := google.CredentialsFromJSON(ctx, p.credentials.ServiceAccountJSON, gcpCloudPlatformReadOnlyScope)
	if err != nil {
		return nil, fmt.Errorf("cannot parse gcp service account credentials: %w", err)
	}

	return creds, nil
}

// CRM builds a *cloudresourcemanager.Service pinned to the record's
// credentials. Returns the raw SDK service so callers can use
// project/org/folder helpers directly.
func (p *GCPProvider) CRM(ctx context.Context) (*cloudresourcemanager.Service, error) {
	opts := []option.ClientOption{
		option.WithCredentialsJSON(p.credentials.ServiceAccountJSON),
		option.WithScopes(gcpCloudPlatformReadOnlyScope),
	}
	if p.httpClient != nil {
		opts = append(opts, option.WithHTTPClient(p.httpClient))
	}

	svc, err := cloudresourcemanager.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot build gcp cloudresourcemanager service: %w", err)
	}

	return svc, nil
}

// CloudAsset builds a *cloudasset.Service pinned to the record's
// credentials. Used by org-scope drivers for transitive IAM-binding
// enumeration via SearchAllIamPolicies.
func (p *GCPProvider) CloudAsset(ctx context.Context) (*cloudasset.Service, error) {
	opts := []option.ClientOption{
		option.WithCredentialsJSON(p.credentials.ServiceAccountJSON),
		option.WithScopes(gcpCloudPlatformReadOnlyScope),
	}
	if p.httpClient != nil {
		opts = append(opts, option.WithHTTPClient(p.httpClient))
	}

	svc, err := cloudasset.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot build gcp cloudasset service: %w", err)
	}

	return svc, nil
}

// Probe verifies the service-account key against the record's scope:
// project scope calls cloudresourcemanager.Projects.Get; organization
// scope calls Organizations.Get. Errors are mapped to typed package
// sentinels via MapSDKError.
func (p *GCPProvider) Probe(ctx context.Context) error {
	switch p.credentials.ScopeKind {
	case coredata.CloudAccountScopeKindGCPProject:
		return p.probeProject(ctx)
	case coredata.CloudAccountScopeKindGCPOrganization:
		return p.probeOrganization(ctx)
	default:
		return fmt.Errorf("cannot probe gcp cloud account: unsupported scope kind %q", p.credentials.ScopeKind)
	}
}

func (p *GCPProvider) probeProject(ctx context.Context) error {
	if p.crmProjects != nil {
		if _, err := p.crmProjects.Get(ctx, p.record.ScopeIdentifier); err != nil {
			return MapSDKError(fmt.Errorf("cannot probe gcp project: %w", err))
		}

		return nil
	}

	svc, err := p.CRM(ctx)
	if err != nil {
		return err
	}

	if _, err := svc.Projects.Get(p.record.ScopeIdentifier).Context(ctx).Do(); err != nil {
		return MapSDKError(fmt.Errorf("cannot probe gcp project: %w", err))
	}

	return nil
}

func (p *GCPProvider) probeOrganization(ctx context.Context) error {
	name := fmt.Sprintf("organizations/%s", p.record.ScopeIdentifier)

	if p.crmOrganizations != nil {
		if _, err := p.crmOrganizations.Get(ctx, name); err != nil {
			return MapSDKError(fmt.Errorf("cannot probe gcp organization: %w", err))
		}

		return nil
	}

	svc, err := p.CRM(ctx)
	if err != nil {
		return err
	}

	if _, err := svc.Organizations.Get(name).Context(ctx).Do(); err != nil {
		return MapSDKError(fmt.Errorf("cannot probe gcp organization: %w", err))
	}

	return nil
}

// Record returns the underlying CloudAccountRecord. Useful for
// drivers that need to discriminate on ScopeKind without
// re-threading the value through their constructor.
func (p *GCPProvider) Record() CloudAccountRecord {
	return p.record
}
