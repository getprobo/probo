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

	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
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
