import { Skeleton } from "@probo/ui";
import { Suspense, useCallback } from "react";
import { useQueryLoader } from "react-relay";
import { useLocation } from "react-router";

import type { ViewerMembershipLayoutQuery } from "#/__generated__/iam/ViewerMembershipLayoutQuery.graphql";
import { useAssume } from "#/hooks/iam/useAssume";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import {
  ViewerMembershipLayout,
  viewerMembershipLayoutQuery,
} from "../../iam/organizations/ViewerMembershipLayout";

function EmployeeLayoutQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<ViewerMembershipLayoutQuery>(
    viewerMembershipLayoutQuery,
  );
  const location = useLocation();

  const onAssumeSuccess = useCallback(
    () =>
      loadQuery({
        organizationId,
        hideSidebar: false,
      }),
    [loadQuery, organizationId],
  );

  useAssume({
    afterAssumePath: location.pathname,
    onSuccess: onAssumeSuccess,
  });

  if (!queryRef) {
    return <Skeleton className="w-full h-screen" />;
  }

  return (
    <Suspense fallback={<Skeleton className="w-full h-screen" />}>
      <ViewerMembershipLayout queryRef={queryRef} hideSidebar />
    </Suspense>
  );
}

export default function EmployeeLayoutLoader() {
  return (
    <IAMRelayProvider>
      <EmployeeLayoutQueryLoader />
    </IAMRelayProvider>
  );
}
