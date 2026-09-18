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

package drivers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions/v2"
	"go.gearno.de/kit/log"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
)

// azureNameResolver names the connected subscription for the source-name
// worker. The official display name is preferred, then the subscription ID
// so two sources in one organization still differ.
type azureNameResolver struct {
	session *cloudazure.Session
	logger  *log.Logger
}

func NewAzureNameResolver(session *cloudazure.Session, logger *log.Logger) NameResolver {
	return &azureNameResolver{session: session, logger: logger}
}

func (r *azureNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	name, err := r.subscriptionDisplayName(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return "", err
		}

		r.logger.WarnCtx(
			ctx,
			"cannot read azure subscription name, using subscription id",
			cloudazure.SafeLogFields(err)...,
		)
	}

	if name := strings.TrimSpace(name); name != "" {
		return name, nil
	}

	return r.session.AccountID(), nil
}

func (r *azureNameResolver) subscriptionDisplayName(ctx context.Context) (string, error) {
	client, err := armsubscriptions.NewClient(
		r.session.TokenCredential(),
		r.session.ARMClientOptions(),
	)
	if err != nil {
		return "", fmt.Errorf("cannot create azure subscriptions client: %w", err)
	}

	resp, err := client.Get(ctx, r.session.AccountID(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot get azure subscription: %w", err)
	}

	if resp.DisplayName == nil {
		return "", nil
	}

	return *resp.DisplayName, nil
}
