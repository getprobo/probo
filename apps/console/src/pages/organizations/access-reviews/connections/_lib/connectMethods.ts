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

import type { ConnectorProtocol } from "#/__generated__/core/AccessReviewSourceProviderListItem_provider.graphql";

export type ConnectMethod = ConnectorProtocol | "CLIENT_CREDENTIALS";

interface ConnectMethodSupport {
  configuredProtocols: ReadonlyArray<ConnectorProtocol>;
  apiKeySupported: boolean;
  apiKeyManaged: boolean;
  clientCredentialsSupported: boolean;
}

const connectMethodPreference: ReadonlyArray<ConnectMethod> = [
  "WORKLOAD_IDENTITY",
  "GITHUB_APP",
  "OAUTH2",
  "API_KEY",
  "CLIENT_CREDENTIALS",
];

export function connectMethods({
  configuredProtocols,
  apiKeySupported,
  apiKeyManaged,
  clientCredentialsSupported,
}: ConnectMethodSupport): ReadonlyArray<ConnectMethod> {
  const supportedMethods = new Set<ConnectMethod>(
    configuredProtocols.filter(
      protocol =>
        protocol === "WORKLOAD_IDENTITY"
        || protocol === "GITHUB_APP"
        || protocol === "OAUTH2",
    ),
  );

  if (apiKeySupported || apiKeyManaged) {
    supportedMethods.add("API_KEY");
  }
  if (clientCredentialsSupported) {
    supportedMethods.add("CLIENT_CREDENTIALS");
  }

  return connectMethodPreference.filter(method => supportedMethods.has(method));
}
