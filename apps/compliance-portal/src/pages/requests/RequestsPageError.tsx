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

import { isRouteErrorResponse, Navigate, useRouteError } from "react-router";

import { GlobalError } from "#/components/errors/GlobalError";
import { gateRedirectPath } from "#/lib/auth/continueUrl";
import { NotFoundError } from "#/lib/relay/errors";

function isNotFound(error: unknown): boolean {
  if (error instanceof NotFoundError) {
    return true;
  }
  return isRouteErrorResponse(error) && error.status === 404;
}

// Route ErrorBoundary for /requests. Renders inside MainLayout's Outlet so the
// TopBar and footer stay visible. Matches Figma "Page / 404 / App shell".
export function RequestsPageError() {
  const error = useRouteError();

  const gateRedirect = gateRedirectPath(error, window.location.href);
  if (gateRedirect) {
    return <Navigate replace to={gateRedirect} />;
  }

  if (isNotFound(error)) {
    return <GlobalError error={new NotFoundError()} />;
  }

  return (
    <GlobalError
      error={error}
      onRetry={() => window.location.reload()}
    />
  );
}
