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

// A driver is one provider's gateway: it holds the credential and every call
// the access review makes against that provider. Listing accounts is the only
// thing every driver can do, so Driver is the type the registry hands back.
//
// Anything a provider may or may not support is an optional interface a caller
// asks the driver for, the way an http.ResponseWriter is asked whether it
// flushes. That keeps a capability the registry never learns about: adding one
// touches the drivers that have it and the caller that wants it, and nothing
// else.
package drivers

import (
	"context"

	"go.probo.inc/probo/pkg/connector/provider"
)

// OrganizationLister is implemented by a driver whose provider can enumerate
// the orgs, workspaces, or teams a connection may be scoped to, for the picker
// UI.
//
// A driver whose scope is instead captured during the OAuth callback
// (PagerDuty's subdomain, Vercel's team, Datadog's domain, Zendesk's subdomain)
// implements nothing here, and the picker correctly shows nothing.
type OrganizationLister interface {
	ListOrganizations(ctx context.Context) ([]Organization, error)
}

// Organizations asks a driver to list the orgs a connection may be scoped to,
// reporting an empty list for a driver with no picker rather than an error —
// having no organizations to choose from is a normal state, not a failure.
func Organizations(ctx context.Context, driver Driver) ([]Organization, error) {
	lister, ok := driver.(OrganizationLister)
	if !ok {
		return nil, nil
	}

	return lister.ListOrganizations(ctx)
}

// InstanceName asks a driver what its provider instance is called, so a source
// can be labelled "GitHub acme" rather than just "GitHub". A driver that cannot
// answer returns "", which keeps the generic name.
func InstanceName(ctx context.Context, driver Driver) (string, error) {
	resolver, ok := driver.(NameResolver)
	if !ok {
		return "", nil
	}

	return resolver.ResolveInstanceName(ctx)
}

// organizationListerFunc lets a plain listing function satisfy
// OrganizationLister without a named type of its own.
type organizationListerFunc func(context.Context) ([]Organization, error)

func (f organizationListerFunc) ListOrganizations(ctx context.Context) ([]Organization, error) {
	return f(ctx)
}

// capable pairs a driver with the optional capabilities its provider supports.
//
// A nil resolver or lister is dropped rather than embedded, so a caller's type
// assertion answers truthfully: embedding a nil interface would satisfy the
// assertion and then panic on the call.
func capable(driver Driver, resolver NameResolver, lister OrganizationLister) Driver {
	switch {
	case resolver != nil && lister != nil:
		return struct {
			Driver
			NameResolver
			OrganizationLister
		}{driver, resolver, lister}

	case resolver != nil:
		return struct {
			Driver
			NameResolver
		}{driver, resolver}

	case lister != nil:
		return struct {
			Driver
			OrganizationLister
		}{driver, lister}
	}

	return driver
}

// organizationsBase is the host an organization lister targets: the provider's
// API root, falling back to Identity for a provider with no static data root of
// its own (DocuSign). It matters that this matches what the driver uses, so an
// endpoint override reaches the picker the same way it reaches the review.
func organizationsBase(endpoints provider.Endpoints) string {
	if endpoints.APIBase != "" {
		return endpoints.APIBase
	}

	return endpoints.Identity
}
