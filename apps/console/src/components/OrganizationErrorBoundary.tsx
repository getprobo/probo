import { AssumptionRequiredError, UnAuthenticatedError } from "@probo/relay";
import { Navigate, useLocation, useRouteError } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";

import { PageError } from "./PageError";

export function OrganizationErrorBoundary() {
  const error = useRouteError();
  const location = useLocation();
  const organizationId = useOrganizationId();

  if (error instanceof UnAuthenticatedError) {
    return <Navigate to="/auth/login" state={{ from: location.pathname }} />;
  }

  if (error instanceof AssumptionRequiredError) {
    return <Navigate to="/auth/login" state={{ from: location.pathname, organizationId }} />;
  }

  return <PageError error={error instanceof Error ? error : new Error("unknown error")} />;
}
