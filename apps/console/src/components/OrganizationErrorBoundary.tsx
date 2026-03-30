// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
