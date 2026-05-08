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
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"

	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// fakeCRMProjects implements the unexported crmProjectsAPI seam
	// declared in gcp.go. It records the requested project id so tests
	// can assert the correct scope identifier was forwarded.
	fakeCRMProjects struct {
		err         error
		calls       int
		lastProject string
	}

	// fakeCRMOrganizations implements the unexported crmOrganizationsAPI
	// seam.
	fakeCRMOrganizations struct {
		err      error
		calls    int
		lastName string
	}
)

func (f *fakeCRMProjects) Get(_ context.Context, projectID string) (*cloudresourcemanager.Project, error) {
	f.calls++
	f.lastProject = projectID
	if f.err != nil {
		return nil, f.err
	}
	return &cloudresourcemanager.Project{ProjectId: projectID}, nil
}

func (f *fakeCRMOrganizations) Get(_ context.Context, name string) (*cloudresourcemanager.Organization, error) {
	f.calls++
	f.lastName = name
	if f.err != nil {
		return nil, f.err
	}
	return &cloudresourcemanager.Organization{Name: name}, nil
}

func newGCPProviderForTest(
	scope coredata.CloudAccountScopeKind,
	scopeIdentifier string,
	projects *fakeCRMProjects,
	orgs *fakeCRMOrganizations,
) *GCPProvider {
	creds := &GCPCredentials{
		ServiceAccountJSON: []byte(`{"type":"service_account","project_id":"probo-test"}`),
		ScopeKind:          scope,
		ProjectID:          "probo-test",
	}
	if scope == coredata.CloudAccountScopeKindGCPOrganization {
		creds.OrganizationID = scopeIdentifier
	}

	rec := CloudAccountRecord{
		ID:              "01HXYZ-cloud-account-gcp",
		Provider:        coredata.CloudAccountProviderGCP,
		Kind:            coredata.CloudAccountCredentialKindGCPServiceAccountKey,
		ScopeKind:       scope,
		ScopeIdentifier: scopeIdentifier,
	}

	p := newGCPProvider(&http.Client{}, rec, creds)
	p.crmProjects = projects
	p.crmOrganizations = orgs

	return p
}

func TestGCPProvider_Probe_ProjectScope(t *testing.T) {
	t.Parallel()

	t.Run("calls Projects.Get with the scope identifier", func(t *testing.T) {
		t.Parallel()

		projects := &fakeCRMProjects{}
		orgs := &fakeCRMOrganizations{}
		p := newGCPProviderForTest(
			coredata.CloudAccountScopeKindGCPProject,
			"probo-test-project",
			projects,
			orgs,
		)

		err := p.Probe(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, projects.calls, "Projects.Get must be called once")
		assert.Equal(t, "probo-test-project", projects.lastProject, "scope identifier must be forwarded verbatim")
		assert.Equal(t, 0, orgs.calls, "Organizations.Get must NOT be called for project scope")
	})

	t.Run("googleapi 401 maps to ErrCredentialsInvalid", func(t *testing.T) {
		t.Parallel()

		projects := &fakeCRMProjects{err: &googleapi.Error{Code: 401, Message: "unauthenticated"}}
		p := newGCPProviderForTest(
			coredata.CloudAccountScopeKindGCPProject,
			"probo-test-project",
			projects,
			&fakeCRMOrganizations{},
		)

		err := p.Probe(context.Background())
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCredentialsInvalid)
	})

	t.Run("googleapi 403 maps to ErrInsufficientPermissions", func(t *testing.T) {
		t.Parallel()

		projects := &fakeCRMProjects{err: &googleapi.Error{Code: 403, Message: "forbidden"}}
		p := newGCPProviderForTest(
			coredata.CloudAccountScopeKindGCPProject,
			"probo-test-project",
			projects,
			&fakeCRMOrganizations{},
		)

		err := p.Probe(context.Background())
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientPermissions)
	})

	t.Run("googleapi 404 maps to ErrScopeUnreachable", func(t *testing.T) {
		t.Parallel()

		projects := &fakeCRMProjects{err: &googleapi.Error{Code: 404, Message: "missing"}}
		p := newGCPProviderForTest(
			coredata.CloudAccountScopeKindGCPProject,
			"probo-test-project",
			projects,
			&fakeCRMOrganizations{},
		)

		err := p.Probe(context.Background())
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrScopeUnreachable)
	})
}

func TestGCPProvider_Probe_OrganizationScope(t *testing.T) {
	t.Parallel()

	t.Run("calls Organizations.Get with organizations/<id>", func(t *testing.T) {
		t.Parallel()

		projects := &fakeCRMProjects{}
		orgs := &fakeCRMOrganizations{}
		p := newGCPProviderForTest(
			coredata.CloudAccountScopeKindGCPOrganization,
			"123456789",
			projects,
			orgs,
		)

		err := p.Probe(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, orgs.calls, "Organizations.Get must be called once")
		assert.Equal(t, fmt.Sprintf("organizations/%s", "123456789"), orgs.lastName, "name must be prefixed with organizations/")
		assert.Equal(t, 0, projects.calls, "Projects.Get must NOT be called for organization scope")
	})

	t.Run("googleapi error chain is mapped via MapSDKError", func(t *testing.T) {
		t.Parallel()

		orgs := &fakeCRMOrganizations{err: &googleapi.Error{Code: 403, Message: "forbidden"}}
		p := newGCPProviderForTest(
			coredata.CloudAccountScopeKindGCPOrganization,
			"123456789",
			&fakeCRMProjects{},
			orgs,
		)

		err := p.Probe(context.Background())
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientPermissions)
	})
}

func TestGCPProvider_Probe_UnsupportedScope(t *testing.T) {
	t.Parallel()

	p := newGCPProviderForTest(
		coredata.CloudAccountScopeKind("MYSTERY"),
		"x",
		&fakeCRMProjects{},
		&fakeCRMOrganizations{},
	)

	err := p.Probe(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scope kind")
}

func TestGCPCredentials_RoundTrip(t *testing.T) {
	t.Parallel()

	original := &GCPCredentials{
		ServiceAccountJSON: []byte(`{"type":"service_account","project_id":"probo-test","private_key":"-----BEGIN-----\n..."}`),
		ScopeKind:          coredata.CloudAccountScopeKindGCPProject,
		ProjectID:          "probo-test",
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	got := &GCPCredentials{}
	require.NoError(t, json.Unmarshal(raw, got))
	assert.Equal(t, original.ServiceAccountJSON, got.ServiceAccountJSON, "service-account JSON bytes must round-trip verbatim")
	assert.Equal(t, original.ScopeKind, got.ScopeKind)
	assert.Equal(t, original.ProjectID, got.ProjectID)
}

func TestGCPCredentials_UnmarshalRejectsForeignKind(t *testing.T) {
	t.Parallel()

	envelope := []byte(`{"v":1,"kind":"AWS_ASSUME_ROLE","payload":{}}`)

	got := &GCPCredentials{}
	err := got.UnmarshalJSON(envelope)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCredentialsInvalid))
}

func TestGCPCredentials_EnvelopeIdentity(t *testing.T) {
	t.Parallel()

	c := &GCPCredentials{}
	assert.Equal(t, coredata.CloudAccountProviderGCP, c.Provider())
	assert.Equal(t, coredata.CloudAccountCredentialKindGCPServiceAccountKey, c.Kind())
}

// TestGCPCredentials_EnvelopeRejectsMismatchedProvider asserts the
// cross-provider rejection path performed by UnmarshalCredentials. A
// payload whose envelope kind belongs to GCP must not be accepted under
// provider=AWS.
func TestGCPCredentials_EnvelopeRejectsMismatchedProvider(t *testing.T) {
	t.Parallel()

	gcp := &GCPCredentials{
		ServiceAccountJSON: []byte(`{"type":"service_account"}`),
		ScopeKind:          coredata.CloudAccountScopeKindGCPProject,
	}
	raw, err := json.Marshal(gcp)
	require.NoError(t, err)

	_, err = UnmarshalCredentials(coredata.CloudAccountProviderAWS, raw)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCredentialsInvalid))
}
