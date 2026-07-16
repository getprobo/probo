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

// Maps a caught gate error to the route that resolves it, carrying the current
// URL as a `continue` target so the user returns here once the gate is cleared.
// Returns a router-relative path (basename applied by the router); the target
// page validates the continue URL via getSafeContinueUrl. Returns null for any
// other error so the boundary can fall through to its normal error UI.
export function resolveGateRedirect(error: unknown): string | null {
  const continueUrl = encodeURIComponent(window.location.href);

  if (error instanceof FullNameRequiredError) {
    return `/full-name?continue=${continueUrl}`;
  }

  if (error instanceof NDASignatureRequiredError) {
    return `/nda?continue=${continueUrl}`;
  }

  return null;
}
