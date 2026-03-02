import { Skeleton } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { ViewerMembershipLayoutQuery } from "#/__generated__/iam/ViewerMembershipLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import {
  ViewerMembershipLayout,
  viewerMembershipLayoutQuery,
} from "./ViewerMembershipLayout";

function ViewerMembershipLayoutQueryLoader() {
  const organizationId = useOrganizationId();

  const [queryRef, loadQuery] = useQueryLoader<ViewerMembershipLayoutQuery>(
    viewerMembershipLayoutQuery,
  );

  useEffect(() => {
    loadQuery({ organizationId, hideSidebar: false });
  }, [organizationId, loadQuery]);

  if (!queryRef) {
    return <Skeleton className="w-full h-screen" />;
  }

  return (
    <Suspense fallback={<Skeleton className="w-full h-screen" />}>
      <ViewerMembershipLayout queryRef={queryRef} />
    </Suspense>
  );
}

export default function ViewerMembershipLayoutLoader() {
  return (
    <IAMRelayProvider>
      <ViewerMembershipLayoutQueryLoader />
    </IAMRelayProvider>
  );
}
