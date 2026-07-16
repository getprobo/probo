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

import { FullNameRequiredError, NDASignatureRequiredError } from "@probo/relay";

import { getPathPrefix } from "#/lib/http/pathPrefix";

// Markers appended to a post-auth `continue` URL so the portal fires the pending
// "request access" mutation once the user lands back authenticated. `request-all`
// covers the top-bar "Get Access"; the per-resource markers carry the id of a
// single document / report / file whose access was requested from a locked row.
export const REQUEST_ALL_PARAM = "request-all";
export const REQUEST_DOCUMENT_PARAM = "request-document-id";
export const REQUEST_REPORT_PARAM = "request-report-id";
export const REQUEST_FILE_PARAM = "request-file-id";

// Validates a `continue` target before we navigate to it. Only same-origin URLs
// under the portal's path prefix are accepted; anything else falls back to the
// portal home, so a crafted `?continue=` can never bounce the user off-site.
export function getSafeContinueUrl(param: string | null | undefined): string {
  const prefix = getPathPrefix();
  const fallback = window.location.origin + (prefix || "/");

  if (!param) {
    return fallback;
  }

  try {
    const url = new URL(param, window.location.origin);
    if (
      url.origin === window.location.origin
      && url.pathname.startsWith(`${prefix}/`)
    ) {
      return window.location.origin + url.pathname + url.search;
    }
  } catch {
    return fallback;
  }

  return fallback;
}

// Absolute URL of the current page with the request-all marker set, used as the
// `continue` target so the access request resumes after sign-in.
export function buildRequestAllContinueUrl(): string {
  const url = new URL(window.location.href);
  url.searchParams.set(REQUEST_ALL_PARAM, "true");
  return url.toString();
}

// Absolute URL of the current page with a per-resource marker set, so a single
// document / report / file access request resumes after sign-in.
export function buildRequestAccessContinueUrl(param: string, id: string): string {
  const url = new URL(window.location.href);
  url.searchParams.set(param, id);
  return url.toString();
}

// Maps a caught auth-gate error to the route that resolves it, carrying the
// given `continueUrl` so the user returns here (and any deferred request
// resumes) once the gate is cleared. Returns null for non-gate errors. Shared
// by the route boundaries and the request-access flows so all gate handling
// stays in one place.
export function gateRedirectPath(error: unknown, continueUrl: string): string | null {
  if (error instanceof FullNameRequiredError) {
    return `/full-name?continue=${encodeURIComponent(continueUrl)}`;
  }
  if (error instanceof NDASignatureRequiredError) {
    return `/nda?continue=${encodeURIComponent(continueUrl)}`;
  }
  return null;
}
