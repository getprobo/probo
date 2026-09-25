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

export function integrationListPath(organizationId: string) {
  return `/organizations/${organizationId}/settings/integrations`;
}

export function marketplacePath(organizationId: string) {
  return `${integrationListPath(organizationId)}/marketplace`;
}

export function connectVendorPath(organizationId: string, provider: string) {
  return `${marketplacePath(organizationId)}/${provider.toLowerCase().replaceAll("_", "-")}`;
}

const connectMethodSlug = {
  WORKLOAD_IDENTITY: "workload-identity",
  GITHUB_APP: "github-app",
  INSTALL: "install",
  OAUTH2: "oauth",
  CLIENT_CREDENTIALS: "client-credentials",
  API_KEY: "api-key",
} as const;

export type ConnectVendorMethod = keyof typeof connectMethodSlug;

export function connectMethodFromSlug(slug: string): ConnectVendorMethod | null {
  const match = (Object.entries(connectMethodSlug) as Array<[ConnectVendorMethod, string]>)
    .find(([, value]) => value === slug);
  return match?.[0] ?? null;
}

export function connectVendorMethodPath(
  organizationId: string,
  provider: string,
  method: ConnectVendorMethod,
) {
  return `${connectVendorPath(organizationId, provider)}/${connectMethodSlug[method]}`;
}

export function providerFromSlug(slug: string) {
  return slug.toUpperCase().replaceAll("-", "_");
}
