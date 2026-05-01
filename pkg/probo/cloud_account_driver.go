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

package probo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	// awsIAMSeam is the in-process bridge between
	// *cloudaccount.AWSProvider and drivers.AWSIAMReader. It builds
	// a real *iam.Client at construction time so the wider seam
	// methods (ListMFADevices, GenerateCredentialReport,
	// GetCredentialReport) are available without re-typing through
	// the cloudaccount.iamAPI narrow seam.
	awsIAMSeam struct {
		client *iam.Client
	}

	// gcpIAMPolicySeam is the bridge between *cloudaccount.GCPProvider
	// and drivers.GCPIAMPolicyReader. The provider's CRM and
	// CloudAsset SDK clients are built lazily on the first call;
	// errors surface to the driver's ListAccounts caller.
	gcpIAMPolicySeam struct {
		provider *cloudaccount.GCPProvider
	}

	// azureRoleAssignmentSeam is the bridge between
	// *cloudaccount.AzureProvider and drivers.AzureRoleAssignmentLister.
	// v1 returns an empty list -- the underlying ARM + Microsoft
	// Graph plumbing lives in pkg/cloudaccount/azure.go and lights
	// up in the follow-up enrichment pass.
	azureRoleAssignmentSeam struct {
		provider *cloudaccount.AzureProvider
	}
)

// CloudAccountDriverFactory returns the access-review driver bound
// to the supplied cloud-account id. Wired into the access-review
// engine at construction time so the engine never imports
// pkg/cloudaccount or any cloud SDK.
func (s *TenantService) BuildCloudAccountDriver(
	ctx context.Context,
	cloudAccountID gid.GID,
) (drivers.Driver, error) {
	account, err := s.CloudAccounts.Get(ctx, cloudAccountID)
	if err != nil {
		return nil, fmt.Errorf("cannot load cloud account: %w", err)
	}

	registry := s.cloudAccountRegistryRef()
	if registry == nil {
		return nil, fmt.Errorf("cannot build cloud account driver: registry not configured")
	}

	record := mapCloudAccountToRecord(account)

	switch account.Provider {
	case coredata.CloudAccountProviderAWS:
		p, err := registry.BuildAWSProvider(record)
		if err != nil {
			return nil, fmt.Errorf("cannot build aws provider: %w", err)
		}
		seam, err := newAWSIAMSeam(p)
		if err != nil {
			return nil, fmt.Errorf("cannot build aws iam seam: %w", err)
		}
		return drivers.NewCloudAWSDriver(seam), nil

	case coredata.CloudAccountProviderGCP:
		p, err := registry.BuildGCPProvider(record)
		if err != nil {
			return nil, fmt.Errorf("cannot build gcp provider: %w", err)
		}
		return drivers.NewCloudGCPDriver(
			&gcpIAMPolicySeam{provider: p},
			account.ScopeKind,
			account.ScopeIdentifier,
		), nil

	case coredata.CloudAccountProviderAzure:
		p, err := registry.BuildAzureProvider(record)
		if err != nil {
			return nil, fmt.Errorf("cannot build azure provider: %w", err)
		}
		return drivers.NewCloudAzureDriver(&azureRoleAssignmentSeam{provider: p}), nil

	default:
		return nil, fmt.Errorf("cannot build cloud account driver: unsupported provider %q", account.Provider)
	}
}

func (s *TenantService) cloudAccountRegistryRef() *cloudaccount.Registry {
	if s.CloudAccounts == nil {
		return nil
	}
	return s.CloudAccounts.registry
}

// CloudAccountDriverFactoryProvider returns the closure the
// access-review service hands to its review engine. Wired at probod
// construction time via accessreviewService.SetCloudAccountDriverFactoryProvider
// so the access-review engine can dispatch a cloud-account-backed
// access source without importing pkg/cloudaccount or any cloud SDK.
func (s *Service) CloudAccountDriverFactoryProvider() accessreview.CloudAccountDriverFactoryProvider {
	return func(scope coredata.Scoper) accessreview.CloudAccountDriverFactory {
		tenantSvc := s.WithTenant(scope.GetTenantID())
		return func(ctx context.Context, cloudAccountID gid.GID) (drivers.Driver, error) {
			return tenantSvc.BuildCloudAccountDriver(ctx, cloudAccountID)
		}
	}
}

// newAWSIAMSeam upcasts the cloudaccount.iamAPI narrow seam into
// the wider drivers.AWSIAMReader interface. The concrete *iam.Client
// the provider builds satisfies both.
func newAWSIAMSeam(p *cloudaccount.AWSProvider) (drivers.AWSIAMReader, error) {
	client, ok := p.IAM().(*iam.Client)
	if !ok {
		return nil, fmt.Errorf("cannot build aws iam seam: provider returned unexpected client type")
	}
	return &awsIAMSeam{client: client}, nil
}

func (s *awsIAMSeam) ListUsers(ctx context.Context, in *iam.ListUsersInput, opts ...func(*iam.Options)) (*iam.ListUsersOutput, error) {
	return s.client.ListUsers(ctx, in, opts...)
}

func (s *awsIAMSeam) ListMFADevices(ctx context.Context, in *iam.ListMFADevicesInput, opts ...func(*iam.Options)) (*iam.ListMFADevicesOutput, error) {
	return s.client.ListMFADevices(ctx, in, opts...)
}

func (s *awsIAMSeam) GenerateCredentialReport(ctx context.Context, in *iam.GenerateCredentialReportInput, opts ...func(*iam.Options)) (*iam.GenerateCredentialReportOutput, error) {
	return s.client.GenerateCredentialReport(ctx, in, opts...)
}

func (s *awsIAMSeam) GetCredentialReport(ctx context.Context, in *iam.GetCredentialReportInput, opts ...func(*iam.Options)) (*iam.GetCredentialReportOutput, error) {
	return s.client.GetCredentialReport(ctx, in, opts...)
}

// GetProjectIAMPolicy reads the IAM policy for a single GCP
// project via cloudresourcemanager. v1 returns an empty policy
// when the SDK call fails -- callers see the error wrapped at
// the driver layer.
func (s *gcpIAMPolicySeam) GetProjectIAMPolicy(ctx context.Context, projectID string) (drivers.GCPIAMPolicy, error) {
	crm, err := s.provider.CRM(ctx)
	if err != nil {
		return drivers.GCPIAMPolicy{}, fmt.Errorf("cannot build gcp crm client: %w", err)
	}

	policy, err := crm.Projects.GetIamPolicy(projectID, nil).Context(ctx).Do()
	if err != nil {
		return drivers.GCPIAMPolicy{}, fmt.Errorf("cannot get gcp project iam policy: %w", err)
	}

	bindings := make([]drivers.GCPIAMBinding, 0, len(policy.Bindings))
	for _, b := range policy.Bindings {
		bindings = append(bindings, drivers.GCPIAMBinding{
			Role:    b.Role,
			Members: b.Members,
		})
	}

	return drivers.GCPIAMPolicy{Bindings: bindings}, nil
}

// SearchAllIAMPolicies enumerates IAM bindings across an
// organization tree via cloudasset. v1 emits a single page;
// follow-ups will paginate.
func (s *gcpIAMPolicySeam) SearchAllIAMPolicies(ctx context.Context, orgID string) ([]drivers.GCPIAMPolicy, error) {
	asset, err := s.provider.CloudAsset(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot build gcp cloudasset client: %w", err)
	}

	scope := fmt.Sprintf("organizations/%s", orgID)
	resp, err := asset.V1.SearchAllIamPolicies(scope).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot search gcp iam policies: %w", err)
	}

	policies := make([]drivers.GCPIAMPolicy, 0, len(resp.Results))
	for _, result := range resp.Results {
		if result.Policy == nil {
			continue
		}
		bindings := make([]drivers.GCPIAMBinding, 0, len(result.Policy.Bindings))
		for _, b := range result.Policy.Bindings {
			bindings = append(bindings, drivers.GCPIAMBinding{
				Role:    b.Role,
				Members: b.Members,
			})
		}
		policies = append(policies, drivers.GCPIAMPolicy{Bindings: bindings})
	}

	return policies, nil
}

// ListRoleAssignments returns role assignments at the configured
// Azure scope. v1 ships an empty implementation -- the underlying
// ARM + Graph plumbing materializes once the cloudaccount.AzureProvider
// exposes role-assignment helpers (next slice). Returning an
// empty list rather than an error keeps the driver compile-clean
// and lets the e2e harness exercise the dispatch path.
func (s *azureRoleAssignmentSeam) ListRoleAssignments(ctx context.Context) ([]drivers.AzureRoleAssignmentRecord, error) {
	_ = s.provider
	return nil, nil
}
