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
