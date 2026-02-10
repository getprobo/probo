import { UnAuthenticatedError } from "@probo/relay";
import { Navigate, useRouteError } from "react-router";

import { PageError } from "./PageError";

export function RootErrorBoundary() {
  const error = useRouteError();

  const search = new URLSearchParams([["redirect-path", window.location.href]]);

  if (error instanceof UnAuthenticatedError) {
    return <Navigate to={{ pathname: "/auth/login", search: "?" + search.toString() }} />;
  }

  return <PageError error={error instanceof Error ? error : new Error("unknown error")} />;
}
