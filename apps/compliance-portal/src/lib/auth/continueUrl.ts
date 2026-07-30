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

import {
  FullNameRequiredError,
  NDASignatureRequiredError,
  UnAuthenticatedError,
} from "@probo/relay";

import { localizedPath, resolveUrlLocale, type UrlLocale } from "#/lib/i18n/locale";

// Markers appended to a post-auth `continue` URL so the portal fires the pending
// "request access" mutation once the user lands back authenticated. The
// per-resource markers carry the id of a single document / report / file whose
// access was requested from a locked row.
export const REQUEST_DOCUMENT_PARAM = "request-document-id";
export const REQUEST_REPORT_PARAM = "request-report-id";
export const REQUEST_FILE_PARAM = "request-file-id";
// Marker that re-opens the "New Request" dialog once the user lands back
// authenticated (the data request form gates on sign-in before it opens).
export const NEW_REQUEST_PARAM = "new-request";
// Marker that re-opens the "Subscribe to updates" dialog after sign-in.
export const SUBSCRIBE_PARAM = "subscribe";

// Validates a `continue` target before we navigate to it. Only same-origin URLs
// are accepted; anything else falls back to the portal home, so a crafted
// `?continue=` can never bounce the user off-site.
export function getSafeContinueUrl(param: string | null | undefined): string {
  const localeHome = localizedPath(resolveUrlLocale(), "/");
  const fallback = window.location.origin + localeHome;

  if (!param) {
    return fallback;
  }

  try {
    const url = new URL(param, window.location.origin);
    if (url.origin === window.location.origin) {
      return window.location.origin + url.pathname + url.search;
    }
  } catch {
    return fallback;
  }

  return fallback;
}

// Absolute URL of the current page with a per-resource marker set, so a single
// document / report / file access request resumes after sign-in.
export function buildRequestAccessContinueUrl(param: string, id: string): string {
  const url = new URL(window.location.href);
  url.searchParams.set(param, id);
  return url.toString();
}

// Absolute URL of the current page with the new-request marker set, so the data
// request dialog re-opens after sign-in.
export function buildNewRequestContinueUrl(): string {
  const url = new URL(window.location.href);
  url.searchParams.set(NEW_REQUEST_PARAM, "true");
  return url.toString();
}

// Absolute URL of the current page with the subscribe marker set, so the
// mailing-list subscribe dialog re-opens after sign-in.
export function buildSubscribeContinueUrl(): string {
  const url = new URL(window.location.href);
  url.searchParams.set(SUBSCRIBE_PARAM, "true");
  return url.toString();
}

// Maps a caught auth-gate error to the route that resolves it, carrying the
// given `continueUrl` so the user returns here (and any deferred request
// resumes) once the gate is cleared. Returns null for non-gate errors. Shared
// by the route boundaries and consumeAuthGate (useMutation) so all gate
// handling stays in one place.
function isGateError(error: unknown, ctor: new (...args: never[]) => Error, name: string): boolean {
  return error instanceof ctor || (error instanceof Error && error.name === name);
}

export function gateRedirectPath(
  error: unknown,
  continueUrl: string,
  locale: UrlLocale = resolveUrlLocale(),
): string | null {
  // Prefer instanceof; fall back to `name` so a duplicated @probo/relay copy
  // (or a wrapped error that preserved the name) still redirects.
  if (isGateError(error, FullNameRequiredError, "FullNameRequiredError")) {
    return `${localizedPath(locale, "/full-name")}?continue=${encodeURIComponent(continueUrl)}`;
  }
  if (isGateError(error, NDASignatureRequiredError, "NDASignatureRequiredError")) {
    return `${localizedPath(locale, "/nda")}?continue=${encodeURIComponent(continueUrl)}`;
  }
  return null;
}

// Sends the browser to the OAuth entry point, carrying a validated continue URL
// so the user returns to the portal (and any deferred access request resumes)
// after sign-in.
export function redirectToInitiate(continueTo: string): void {
  const initiateURL = new URL("/initiate", window.location.origin);
  initiateURL.searchParams.set("continue", getSafeContinueUrl(continueTo));
  window.location.href = initiateURL.toString();
}

// Consumes an auth-gate error from a mutation (or similar async path): redirects
// to OAuth /initiate, the full-name page, or the NDA page as appropriate.
// Returns true when the error was a gate and navigation was kicked off, so the
// caller can skip toasts / local error UI. Used by the portal's useMutation
// binding so every mutation gets this for free.
export function consumeAuthGate(
  error: unknown,
  continueUrl: string,
  navigate: (to: string) => void,
  locale: UrlLocale = resolveUrlLocale(),
): boolean {
  if (isGateError(error, UnAuthenticatedError, "UnAuthenticatedError")) {
    redirectToInitiate(continueUrl);
    return true;
  }

  const gatePath = gateRedirectPath(error, continueUrl, locale);
  if (gatePath) {
    navigate(gatePath);
    return true;
  }

  return false;
}
