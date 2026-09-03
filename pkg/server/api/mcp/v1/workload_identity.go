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

package mcp_v1

import (
	"context"
	"encoding/json"
	"fmt"

	"go.gearno.de/kit/log"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func (r *Resolver) workloadIdentitySettings(
	ctx context.Context,
	input *types.CreateWorkloadIdentityConnectorInput,
) ([]byte, error) {
	switch input.Provider {
	case coredata.ConnectorProviderAWS:
		if input.AwsRoleArn == nil || *input.AwsRoleArn == "" {
			return nil, fmt.Errorf("aws_role_arn is required")
		}

		settings, err := cloudaws.NewConnectorSettings(*input.AwsRoleArn)
		if err != nil {
			return nil, err
		}

		raw, err := json.Marshal(settings)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot marshal aws connector settings", log.Error(err))

			return nil, fmt.Errorf("internal server error")
		}

		return raw, nil
	case coredata.ConnectorProviderGCP:
		if input.GcpWorkloadIdentityProvider == nil ||
			*input.GcpWorkloadIdentityProvider == "" ||
			input.GcpServiceAccountEmail == nil ||
			*input.GcpServiceAccountEmail == "" {
			return nil, fmt.Errorf("gcp_workload_identity_provider and gcp_service_account_email are required")
		}

		validated, err := cloudgcp.NewConnectorSettings(
			*input.GcpWorkloadIdentityProvider,
			*input.GcpServiceAccountEmail,
		)
		if err != nil {
			return nil, err
		}

		settings := coredata.GCPConnectorSettings{
			WorkloadIdentityProvider: validated.WorkloadIdentityProvider,
			ServiceAccountEmail:      validated.ServiceAccountEmail,
		}

		raw, err := json.Marshal(settings)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot marshal gcp connector settings", log.Error(err))

			return nil, fmt.Errorf("internal server error")
		}

		return raw, nil
	default:
		return nil, fmt.Errorf("provider does not support workload identity")
	}
}
