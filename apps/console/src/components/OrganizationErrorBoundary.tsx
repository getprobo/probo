import { AssumptionRequiredError, UnAuthenticatedError } from "@probo/relay";
import { Navigate, useRouteError } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";

import { PageError } from "./PageError";

export function OrganizationErrorBoundary() {
  const error = useRouteError();
  const organizationId = useOrganizationId();

  const search = new URLSearchParams([
    ["organization-id", organizationId],
    ["redirect-path", window.location.pathname + window.location.search],
  ]);

  if (error instanceof UnAuthenticatedError) {
    return <Navigate to={{ pathname: "/auth/login", search: "?" + search.toString() }} />;
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
