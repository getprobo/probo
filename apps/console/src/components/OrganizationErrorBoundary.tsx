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
import { Navigate, useLocation, useRouteError } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";

import { PageError } from "./PageError";

export function OrganizationErrorBoundary() {
  const error = useRouteError();
  const organizationId = useOrganizationId();
  const location = useLocation();

  const search = new URLSearchParams();

  if (location.pathname !== "/" || location.search !== "") {
    search.set("continue", window.location.href);
  }

  const queryString = search.toString();

  if (error instanceof UnAuthenticatedError) {
    return <Navigate to={{ pathname: "/auth/login", search: queryString ? "?" + queryString : "" }} />;
  }

  if (error instanceof AssumptionRequiredError) {
    return (
      <Navigate
        to={{
          pathname: `/organizations/${organizationId}/assume`,
          search: "?" + search.toString(),
        }}
      />
    );
  }

  return <PageError error={error instanceof Error ? error : new Error("unknown error")} />;
}
