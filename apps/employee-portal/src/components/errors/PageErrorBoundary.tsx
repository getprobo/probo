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

import { AssumptionRequiredError, UnAuthenticatedError } from "@probo/relay";
import { Suspense } from "react";
import { useParams, useRouteError } from "react-router";

import { redirectToLogin } from "#/lib/auth/redirectToLogin";
import { IAMRelayProvider } from "#/lib/relay/IAMRelayProvider";
import { AssumeOrganizationSession } from "#/pages/iam/_components/errors/AssumeOrganizationSession";
import { MainLayoutSkeleton } from "#/pages/iam/MainLayoutSkeleton";

import { GlobalError } from "./GlobalError";

// Child-route boundary: a page failure is contained to the layout's Outlet, so
// the error renders inside the app chrome (TopBar survives).
export function PageErrorBoundary() {
  const error = useRouteError();
  const { organizationId } = useParams();

  if (error instanceof UnAuthenticatedError) {
    redirectToLogin({ organizationId });
    return null;
  }

  if (error instanceof AssumptionRequiredError) {
    return (
      <IAMRelayProvider>
        <Suspense fallback={<MainLayoutSkeleton />}>
          <AssumeOrganizationSession />
        </Suspense>
      </IAMRelayProvider>
    );
  }

  return (
    <GlobalError
      error={error}
      onRetry={() => window.location.reload()}
    />
  );
}
