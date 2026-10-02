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

import type { ConnectorProtocol } from "#/__generated__/core/ConnectorProviderListItem_provider.graphql";

import { buildConnectorInitiateURL } from "../../../_lib/connectorInitiate";

export function connectOAuthProvider(
  organizationId: string,
  provider: string,
  oauth2Scopes: ReadonlyArray<string>,
  extras?: Record<string, string>,
  name?: string,
) {
  connectProviderProtocol(organizationId, provider, "OAUTH2", {
    oauth2Scopes,
    extras,
    name,
  });
}

export function connectProviderInstall(
  organizationId: string,
  provider: string,
  name: string,
) {
  window.location.assign(installInitiateURL(organizationId, provider, name));
}

export function connectProviderProtocol(
  organizationId: string,
  provider: string,
  protocol: ConnectorProtocol,
  options?: {
    oauth2Scopes?: ReadonlyArray<string>;
    connectorId?: string;
    extras?: Record<string, string>;
    name?: string;
  },
) {
  window.location.assign(
    buildConnectorInitiateURL(organizationId, provider, protocol, options),
  );
}

// No continue parameter: the vendor redirects to Probo's callback, which
// rebuilds the connections URL server-side.
function installInitiateURL(
  organizationId: string,
  provider: string,
  name?: string,
): string {
  const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/connectors/install/initiate", baseURL);
  url.searchParams.append("organization_id", organizationId);
  url.searchParams.append("provider", provider);
  if (name) {
    url.searchParams.append("name", name);
  }
  return url.toString();
}
