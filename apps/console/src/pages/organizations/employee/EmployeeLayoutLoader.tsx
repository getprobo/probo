import { Skeleton } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { ViewerMembershipLayoutQuery } from "#/__generated__/iam/ViewerMembershipLayoutQuery.graphql";
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

  useEffect(() => {
    loadQuery({ organizationId, hideSidebar: true });
  }, [organizationId, loadQuery]);

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
