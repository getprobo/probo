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

package probo

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	// WorkloadIdentitySettingsInput is the provider-specific fields needed to
	// persist a workload-identity connector. Surfaces extract these from their
	// own input types.
	WorkloadIdentitySettingsInput struct {
		Provider                    coredata.ConnectorProvider
		AWSRoleARN                  string
		GCPWorkloadIdentityProvider string
		GCPServiceAccountEmail      string
		AzureTenantID               string
		AzureClientID               string
		AzureSubscriptionID         string
		AzureEnvironment            cloudazure.Environment
	}
)

var (
	// ErrMarshalWorkloadIdentitySettings is returned when validated settings
	// cannot be serialized. Surfaces must map it to an opaque internal error.
	ErrMarshalWorkloadIdentitySettings = errors.New("cannot marshal workload identity settings")
)

// MarshalWorkloadIdentitySettings validates provider-specific workload-identity
// fields and returns the JSON persisted on Connector.RawSettings.
func MarshalWorkloadIdentitySettings(input WorkloadIdentitySettingsInput) ([]byte, error) {
	switch input.Provider {
	case coredata.ConnectorProviderAWS:
		if input.AWSRoleARN == "" {
			return nil, fmt.Errorf("awsRoleArn is required")
		}

		settings, err := cloudaws.NewConnectorSettings(input.AWSRoleARN)
		if err != nil {
			return nil, err
		}

		return marshalWorkloadIdentitySettings(settings)
	case coredata.ConnectorProviderGCP:
		if input.GCPWorkloadIdentityProvider == "" || input.GCPServiceAccountEmail == "" {
			return nil, fmt.Errorf("gcpWorkloadIdentityProvider and gcpServiceAccountEmail are required")
		}

		validated, err := cloudgcp.NewConnectorSettings(
			input.GCPWorkloadIdentityProvider,
			input.GCPServiceAccountEmail,
		)
		if err != nil {
			return nil, err
		}

		settings := coredata.GCPConnectorSettings{
			WorkloadIdentityProvider: validated.WorkloadIdentityProvider,
			ServiceAccountEmail:      validated.ServiceAccountEmail,
		}

		return marshalWorkloadIdentitySettings(settings)
	case coredata.ConnectorProviderAzure:
		if input.AzureTenantID == "" || input.AzureClientID == "" || input.AzureSubscriptionID == "" {
			return nil, fmt.Errorf("azureTenantId, azureClientId and azureSubscriptionId are required")
		}

		validated, err := cloudazure.NewConnectorSettings(
			input.AzureTenantID,
			input.AzureClientID,
			input.AzureSubscriptionID,
			string(input.AzureEnvironment),
		)
		if err != nil {
			return nil, err
		}

		settings := coredata.AzureConnectorSettings{
			TenantID:       validated.TenantID,
			ClientID:       validated.ClientID,
			SubscriptionID: validated.SubscriptionID,
			Environment:    string(validated.Environment),
		}

		return marshalWorkloadIdentitySettings(settings)
	default:
		return nil, fmt.Errorf("provider does not support workload identity")
	}
}

func marshalWorkloadIdentitySettings(settings any) ([]byte, error) {
	raw, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMarshalWorkloadIdentitySettings, err)
	}

	return raw, nil
}

// OrganizationSettingsInput is the organization-connect form: one credential
// for a whole cloud organization rather than a single account.
type OrganizationSettingsInput struct {
	Provider coredata.ConnectorProvider

	// AWSRoleARN names the management account's role, which is where
	// organizations:ListAccounts is granted. AWSMemberRoleName is the role
	// the customer's StackSet created in every member; empty means the name
	// the published template uses.
	AWSRoleARN        string
	AWSMemberRoleName string

	// GCPParent is the Cloud Asset scope discovery and reads run under. It is
	// the confinement boundary, so it is required: without it discovery would
	// be every project the service account happens to see.
	GCPWorkloadIdentityProvider string
	GCPServiceAccountEmail      string
	GCPParent                   string

	// Azure names no subscription here: discovery is what finds them.
	AzureTenantID    string
	AzureClientID    string
	AzureEnvironment string
}

// MarshalOrganizationSettings validates the organization-connect fields and
// returns the JSON persisted on Connector.RawSettings.
func MarshalOrganizationSettings(input OrganizationSettingsInput) ([]byte, error) {
	switch input.Provider {
	case coredata.ConnectorProviderAWS:
		if input.AWSRoleARN == "" {
			return nil, fmt.Errorf("awsRoleArn is required")
		}

		settings, err := cloudaws.NewConnectorSettings(input.AWSRoleARN)
		if err != nil {
			return nil, err
		}

		if input.AWSMemberRoleName != "" {
			if !awsRoleNamePattern.MatchString(input.AWSMemberRoleName) {
				return nil, fmt.Errorf("awsMemberRoleName is not a valid IAM role name")
			}

			settings.MemberRoleName = input.AWSMemberRoleName
		}

		return marshalWorkloadIdentitySettings(settings)
	case coredata.ConnectorProviderGCP:
		if input.GCPWorkloadIdentityProvider == "" || input.GCPServiceAccountEmail == "" {
			return nil, fmt.Errorf("gcpWorkloadIdentityProvider and gcpServiceAccountEmail are required")
		}

		if !gcpParentPattern.MatchString(input.GCPParent) {
			return nil, fmt.Errorf("gcpParent must be organizations/{number} or folders/{number}")
		}

		validated, err := cloudgcp.NewConnectorSettings(
			input.GCPWorkloadIdentityProvider,
			input.GCPServiceAccountEmail,
		)
		if err != nil {
			return nil, err
		}

		settings := coredata.GCPConnectorSettings{
			WorkloadIdentityProvider: validated.WorkloadIdentityProvider,
			ServiceAccountEmail:      validated.ServiceAccountEmail,
			Parent:                   input.GCPParent,
		}

		return marshalWorkloadIdentitySettings(settings)
	case coredata.ConnectorProviderAzure:
		if input.AzureTenantID == "" || input.AzureClientID == "" {
			return nil, fmt.Errorf("azureTenantId and azureClientId are required")
		}

		validated, err := cloudazure.NewConnectorSettings(
			input.AzureTenantID,
			input.AzureClientID,
			"",
			input.AzureEnvironment,
		)
		if err != nil {
			return nil, err
		}

		settings := coredata.AzureConnectorSettings{
			TenantID:    validated.TenantID,
			ClientID:    validated.ClientID,
			Environment: string(validated.Environment),
		}

		return marshalWorkloadIdentitySettings(settings)
	default:
		return nil, fmt.Errorf("provider does not support organization install")
	}
}

var (
	// awsRoleNamePattern is the IAM role-name charset the CloudFormation and
	// Terraform variables already constrain the member role to.
	awsRoleNamePattern = regexp.MustCompile(`^[\w+=,.@-]{1,64}$`)

	// gcpParentPattern is a Cloud Asset scope: an organization or a folder.
	// A project is not accepted — scoping there reproduces exactly the
	// project-only blind spot organization install exists to close.
	gcpParentPattern = regexp.MustCompile(`^(organizations|folders)/[1-9][0-9]*$`)
)
