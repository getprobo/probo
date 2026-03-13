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
