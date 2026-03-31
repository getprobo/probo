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

import { FullNameRequiredError, NDASignatureRequiredError, UnAuthenticatedError } from "@probo/relay";
import { Navigate, useLocation, useRouteError } from "react-router";

import { getPathPrefix } from "#/utils/pathPrefix";

import { PageError } from "./PageError";

export function RootErrorBoundary() {
  const error = useRouteError();
  const location = useLocation();

  const search = new URLSearchParams();

  if (location.pathname !== (getPathPrefix() || "/") || location.search !== "") {
    search.set("continue", window.location.href);
  }

  const queryString = search.toString();

  if (error instanceof UnAuthenticatedError) {
    return (
      <Navigate
        replace
        to={{
          pathname: "/connect",
          search: queryString ? "?" + queryString : "",
        }}
      />
    );
  }

  if (error instanceof FullNameRequiredError) {
    return (
      <Navigate
        replace
        to={{
          pathname: "/full-name",
          search: queryString ? "?" + queryString : "",
        }}
      />
    );
  }

  if (error instanceof NDASignatureRequiredError) {
    return (
      <Navigate
        replace
        to={{
          pathname: "/nda",
          search: queryString ? "?" + queryString : "",
        }}
      />
    );
  }

  return <PageError error={error instanceof Error ? error : new Error("unknown error")} />;
}
