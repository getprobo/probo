import { UnAuthenticatedError } from "@probo/relay";
import { Navigate, useLocation, useRouteError } from "react-router";

import { PageError } from "./PageError";

export function RootErrorBoundary() {
  const error = useRouteError();
  const location = useLocation();

  const search = new URLSearchParams();

  if (location.pathname !== "/" || location.search !== "") {
    search.set("continue", window.location.href);
  }

  const queryString = search.toString();

  if (error instanceof UnAuthenticatedError) {
    return (
      <Navigate to={{
        pathname: "/auth/login",
        search: queryString ? "?" + queryString : "",
      }}
      />
    );
  }

  return <PageError error={error instanceof Error ? error : new Error("unknown error")} />;
}
