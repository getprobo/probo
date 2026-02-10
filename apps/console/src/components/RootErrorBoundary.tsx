import { UnAuthenticatedError } from "@probo/relay";
import { Navigate, useLocation, useRouteError } from "react-router";

import { PageError } from "./PageError";

export function RootErrorBoundary() {
  const error = useRouteError();
  const location = useLocation();

  if (error instanceof UnAuthenticatedError) {
    return <Navigate to="/auth/login" state={{ from: location.pathname }} />;
  }

  return <PageError error={error instanceof Error ? error : new Error("unknown error")} />;
}
