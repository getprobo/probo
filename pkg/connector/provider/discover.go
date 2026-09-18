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

package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions/v2"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	organizationstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"go.probo.inc/probo/pkg/cloud"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
	cloudasset "google.golang.org/api/cloudasset/v1"
)

const (
	// maxDiscoverPages bounds every discovery pager. A customer with more
	// accounts than this has an organization Probo cannot review in one pass,
	// which is a conversation rather than a silent truncation.
	maxDiscoverPages = 50

	// gcpProjectAssetType is the Cloud Asset type of a GCP project.
	gcpProjectAssetType = "cloudresourcemanager.googleapis.com/Project"

	// gcpSearchPageSize is the maximum searchAllResources accepts.
	gcpSearchPageSize = 500
)

// gcpProjectNumberFromAssetName pulls the project number out of a Cloud Asset
// full resource name, which is //cloudresourcemanager.googleapis.com/projects/{number}.
var gcpProjectNumberFromAssetName = regexp.MustCompile(`projects/([1-9][0-9]*)$`)

// discoverAWSAccounts lists the active member accounts of the organization
// the management role belongs to.
//
// organizations:ListAccounts is granted on the management role only — the
// published template gives members the audit role and no Organizations reads
// — so this runs on the connector's own session, never a member's.
func discoverAWSAccounts(
	ctx context.Context,
	session cloud.Session,
	_ *coredata.Connector,
) ([]cloud.Account, error) {
	awsSession, ok := session.(*cloudaws.Session)
	if !ok {
		return nil, fmt.Errorf("cannot discover aws accounts: session is for %s", session.Cloud())
	}

	paginator := organizations.NewListAccountsPaginator(
		organizations.NewFromConfig(awsSession.Config()),
		&organizations.ListAccountsInput{},
	)

	var accounts []cloud.Account

	for range maxDiscoverPages {
		if !paginator.HasMorePages() {
			return accounts, nil
		}

		out, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot list aws organization accounts: %w", err)
		}

		for _, account := range out.Accounts {
			// A suspended account cannot be assumed into, so offering it
			// would only produce a source that fails its first fetch.
			if account.Status != organizationstypes.AccountStatusActive {
				continue
			}

			accounts = append(accounts, cloud.Account{
				ID:   awssdk.ToString(account.Id),
				Name: awssdk.ToString(account.Name),
			})
		}
	}

	return nil, fmt.Errorf("cannot list aws organization accounts: more than %d pages", maxDiscoverPages)
}

// discoverAzureSubscriptions lists the subscriptions the federated
// application can see in its tenant.
func discoverAzureSubscriptions(
	ctx context.Context,
	session cloud.Session,
	_ *coredata.Connector,
) ([]cloud.Account, error) {
	azureSession, ok := session.(*cloudazure.Session)
	if !ok {
		return nil, fmt.Errorf("cannot discover azure subscriptions: session is for %s", session.Cloud())
	}

	client, err := armsubscriptions.NewClient(
		azureSession.TokenCredential(),
		azureSession.ARMClientOptions(),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create azure subscriptions client: %w", err)
	}

	pager := client.NewListPager(nil)

	var accounts []cloud.Account

	for range maxDiscoverPages {
		if !pager.More() {
			return accounts, nil
		}

		out, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot list azure subscriptions: %w", err)
		}

		for _, subscription := range out.Value {
			if subscription == nil || subscription.SubscriptionID == nil {
				continue
			}

			account := cloud.Account{ID: *subscription.SubscriptionID}
			if subscription.DisplayName != nil {
				account.Name = *subscription.DisplayName
			}

			accounts = append(accounts, account)
		}
	}

	return nil, fmt.Errorf("cannot list azure subscriptions: more than %d pages", maxDiscoverPages)
}

// discoverGCPProjects lists the projects under the connector's Parent.
//
// Cloud Asset searchAllResources rather than projects.search: a
// projects.search parent filter matches only the immediate parent, so a
// project nested under a folder is silently omitted — which reads to the
// customer as "Probo cannot see my project". searchAllResources covers
// arbitrary nesting in one paginated call and reuses the permission Cloud
// Asset already needs.
//
// Parent is also the confinement boundary. Without it, discovery would be
// "every project this service account can see", which for a reused or
// long-lived account includes projects the customer never meant to connect.
func discoverGCPProjects(
	ctx context.Context,
	session cloud.Session,
	conn *coredata.Connector,
) ([]cloud.Account, error) {
	gcpSession, ok := session.(*cloudgcp.Session)
	if !ok {
		return nil, fmt.Errorf("cannot discover gcp projects: session is for %s", session.Cloud())
	}

	settings, err := coredata.ConnectorSettings[coredata.GCPConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read gcp connector settings: %w", err)
	}

	if settings.Parent == "" {
		return nil, fmt.Errorf(
			"cannot discover gcp projects: this connector has no organization scope, reconnect it as an organization install",
		)
	}

	svc, err := cloudasset.NewService(ctx, gcpSession.ServiceOptions()...)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcp cloud asset client: %w", err)
	}

	var (
		accounts []cloud.Account
		pages    int
	)

	err = svc.V1.
		SearchAllResources(settings.Parent).
		AssetTypes(gcpProjectAssetType).
		PageSize(gcpSearchPageSize).
		Pages(ctx, func(page *cloudasset.SearchAllResourcesResponse) error {
			pages++
			if pages > maxDiscoverPages {
				return fmt.Errorf("more than %d pages", maxDiscoverPages)
			}

			for _, result := range page.Results {
				// The number, not the id: Cloud Asset full resource names and
				// the _Required log view path are both built from it, so
				// storing the other form would mean converting at every use.
				matches := gcpProjectNumberFromAssetName.FindStringSubmatch(result.Name)
				if len(matches) < 2 {
					continue
				}

				accounts = append(accounts, cloud.Account{
					ID:   matches[1],
					Name: result.DisplayName,
				})
			}

			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("cannot search gcp projects: %w", err)
	}

	return accounts, nil
}
