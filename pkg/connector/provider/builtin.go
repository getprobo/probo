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
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
)

// NewBuiltinRegistry returns a *Registry populated with every connector
// provider compiled into the binary. It panics on duplicate registration or
// invalid Registration metadata — both are programmer errors caught at process
// start, not at runtime.
//
// Deployments that pass endpoint overrides must use NewBuiltinRegistryWith
// instead: an override comes from operator configuration, so a bad one is a
// misconfiguration to report, not a programmer error to crash on.
func NewBuiltinRegistry() *Registry {
	r, err := NewBuiltinRegistryWith()
	if err != nil {
		panic(err)
	}

	return r
}

// NewBuiltinRegistryWith builds the registry with opts applied, returning an
// error rather than panicking so probod can surface a bad endpoint override as
// a startup failure naming the offending provider and field.
func NewBuiltinRegistryWith(opts ...Option) (*Registry, error) {
	var options registryOptions
	for _, opt := range opts {
		opt(&options)
	}

	r := NewRegistry()

	seen := make(map[coredata.ConnectorProvider]bool, len(options.endpoints))

	for _, reg := range []*Registration{
		anthropicRegistration(),
		apolloRegistration(),
		asanaRegistration(),
		betterStackRegistration(),
		bitbucketRegistration(),
		brevoRegistration(),
		brexRegistration(),
		clickhouseRegistration(),
		clickupRegistration(),
		cloudflareRegistration(),
		crispRegistration(),
		cursorRegistration(),
		datadogRegistration(),
		deepgramRegistration(),
		docusignRegistration(),
		dotfileRegistration(),
		grafanaRegistration(),
		githubRegistration(),
		gitlabRegistration(),
		googleAnalyticsRegistration(),
		googleWorkspaceRegistration(),
		herokuRegistration(),
		hubspotRegistration(),
		incidentioRegistration(),
		intercomRegistration(),
		langfuseRegistration(),
		linearRegistration(),
		mercuryRegistration(),
		metabaseRegistration(),
		microsoft365Registration(),
		mondayRegistration(),
		neonRegistration(),
		netlifyRegistration(),
		notionRegistration(),
		nukiRegistration(),
		oktaRegistration(),
		onePasswordRegistration(),
		openaiRegistration(),
		openrouterRegistration(),
		posthogRegistration(),
		pagerdutyRegistration(),
		pylonRegistration(),
		qoveryRegistration(),
		railwayRegistration(),
		renderRegistration(),
		resendRegistration(),
		scalewayRegistration(),
		segmentRegistration(),
		sendgridRegistration(),
		sentryRegistration(),
		signozRegistration(),
		slackRegistration(),
		squareRegistration(),
		supabaseRegistration(),
		tailscaleRegistration(),
		tallyRegistration(),
		upcloudRegistration(),
		vercelRegistration(),
		yousignRegistration(),
		zendeskRegistration(),
	} {
		if o, ok := options.endpoints[reg.Provider]; ok {
			seen[reg.Provider] = true

			endpoints, err := applyEndpointOverride(reg.Provider, reg.EndpointOverrideUnsupported, reg.Endpoints, o)
			if err != nil {
				return nil, err
			}

			reg.Endpoints = endpoints
		}

		if err := r.Register(reg); err != nil {
			return nil, err
		}
	}

	// An override naming a provider that does not exist — a typo, or a provider
	// removed since the config was written — would sit in the config doing
	// nothing while the operator believed the connector was repointed.
	for p := range options.endpoints {
		if !seen[p] {
			return nil, fmt.Errorf("cannot override endpoints for connector provider %q: unknown provider", p)
		}
	}

	return r, nil
}
