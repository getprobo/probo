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
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func (r *Resolver) workloadIdentitySettings(
	ctx context.Context,
	input *types.CreateWorkloadIdentityConnectorInput,
) ([]byte, error) {
	raw, err := probo.MarshalWorkloadIdentitySettings(
		probo.WorkloadIdentitySettingsInput{
			Provider:                    input.Provider,
			AWSRoleARN:                  ref.UnrefOrZero(input.AwsRoleArn),
			GCPWorkloadIdentityProvider: ref.UnrefOrZero(input.GcpWorkloadIdentityProvider),
			GCPServiceAccountEmail:      ref.UnrefOrZero(input.GcpServiceAccountEmail),
			AzureTenantID:               ref.UnrefOrZero(input.AzureTenantID),
			AzureClientID:               ref.UnrefOrZero(input.AzureClientID),
			AzureSubscriptionID:         ref.UnrefOrZero(input.AzureSubscriptionID),
			AzureEnvironment:            ref.UnrefOrZero(input.AzureEnvironment),
		},
	)
	if err != nil {
		if errors.Is(err, probo.ErrMarshalWorkloadIdentitySettings) {
			r.logger.ErrorCtx(ctx, "cannot marshal workload identity connector settings", log.Error(err))

			return nil, fmt.Errorf("internal server error")
		}

		return nil, err
	}

	return raw, nil
}
