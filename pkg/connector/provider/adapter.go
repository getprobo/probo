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

// Every factory on a Registration takes a *Handle, so the registry can call one
// field per capability whatever credential the provider holds. The adapters
// below are what keep that uniform signature from leaking the credential split
// into ~60 provider files: a provider declares the factory it can actually
// write — against an *http.Client, or against a cloud.Session — and the adapter
// unwraps the Handle for it. These functions are the only place that unwrap
// happens.
package provider

import (
	"context"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/coredata"
)

// HTTP adapts a driver factory written against an authenticated *http.Client.
func HTTP(
	fn func(context.Context, *http.Client, *coredata.Connector, *log.Logger, Endpoints) (drivers.Driver, error),
) func(context.Context, *Handle, *log.Logger) (drivers.Driver, error) {
	return func(ctx context.Context, h *Handle, logger *log.Logger) (drivers.Driver, error) {
		httpClient, err := h.http()
		if err != nil {
			return nil, err
		}

		return fn(ctx, httpClient, h.Connector, logger, h.reg.Endpoints)
	}
}

// Cloud adapts a driver factory written against a cloud.Session. A consumer
// needing more than the narrow session interface type-asserts the concrete
// session inside fn; see pkg/cloud.
func Cloud(
	fn func(context.Context, cloud.Session, *coredata.Connector, *log.Logger) (drivers.Driver, error),
) func(context.Context, *Handle, *log.Logger) (drivers.Driver, error) {
	return func(ctx context.Context, h *Handle, logger *log.Logger) (drivers.Driver, error) {
		session, err := h.cloud()
		if err != nil {
			return nil, err
		}

		return fn(ctx, session, h.Connector, logger)
	}
}

// HTTPNameResolver adapts a display-name resolver factory written against an
// authenticated *http.Client. A handle carrying no HTTP credential yields no
// resolver, which keeps the generic source name — the same outcome as a
// provider that registers none.
func HTTPNameResolver(
	fn func(context.Context, *http.Client, *coredata.Connector, *log.Logger, Endpoints) drivers.NameResolver,
) func(context.Context, *Handle, *log.Logger) drivers.NameResolver {
	return func(ctx context.Context, h *Handle, logger *log.Logger) drivers.NameResolver {
		httpClient, err := h.http()
		if err != nil {
			logger.ErrorCtx(ctx, "cannot resolve source name", log.Error(err))

			return nil
		}

		return fn(ctx, httpClient, h.Connector, logger, h.reg.Endpoints)
	}
}

// HTTPProbe adapts a connection check written against an authenticated
// *http.Client.
func HTTPProbe(
	fn func(context.Context, *http.Client, *coredata.Connector, Endpoints) error,
) func(context.Context, *Handle) error {
	return func(ctx context.Context, h *Handle) error {
		httpClient, err := h.http()
		if err != nil {
			return err
		}

		return fn(ctx, httpClient, h.Connector, h.reg.Endpoints)
	}
}

// HTTPOrganizations adapts an organization lister written against an
// authenticated *http.Client.
//
// The lister targets the registration's API root, falling back to Identity for
// a provider with no static data root of its own (DocuSign), so an endpoint
// override reaches the picker the same way it reaches the driver. "" means no
// override is in scope and the lister uses its production base.
func HTTPOrganizations(
	fn func(context.Context, *http.Client, string) ([]drivers.Organization, error),
) func(context.Context, *Handle) ([]drivers.Organization, error) {
	return func(ctx context.Context, h *Handle) ([]drivers.Organization, error) {
		httpClient, err := h.http()
		if err != nil {
			return nil, err
		}

		baseURL := h.reg.Endpoints.APIBase
		if baseURL == "" {
			baseURL = h.reg.Endpoints.Identity
		}

		return fn(ctx, httpClient, baseURL)
	}
}
