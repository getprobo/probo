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

interface LoginRedirectOptions {
  // Absolute URL returned to after login. Defaults to the current page.
  continueUrl?: string;
  // When set, login can resume an organization session (password / SSO).
  organizationId?: string;
}

// Query string for `/auth/login` (and SSO URLs that accept the same params).
export function loginSearch(options: LoginRedirectOptions = {}): URLSearchParams {
  const search = new URLSearchParams();
  search.set("continue", options.continueUrl ?? window.location.href);
  if (options.organizationId != null) {
    search.set("organization-id", options.organizationId);
  }
  return search;
}

// Full navigation to the console login page. `replace` so Back does not return
// to the failing employee-portal page. Never use React Router `Navigate` —
// the router basename would prefix `/employee-portal`.
export function redirectToLogin(options: LoginRedirectOptions = {}): void {
  const url = new URL("/auth/login", window.location.origin);
  for (const [key, value] of loginSearch(options)) {
    url.searchParams.set(key, value);
  }
  window.location.replace(url);
}
