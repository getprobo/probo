import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";
import { Skeleton } from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { ViewerMembershipLayoutQuery } from "/__generated__/iam/ViewerMembershipLayoutQuery.graphql";
import {
  ViewerMembershipLayout,
  viewerMembershipLayoutQuery,
} from "./ViewerMembershipLayout";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

function MembershipLayoutLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<ViewerMembershipLayoutQuery>(
    viewerMembershipLayoutQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
      hideSidebar: false,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <Skeleton className="w-full h-screen" />;
  }

  return (
    <Suspense fallback={<Skeleton className="w-full h-screen" />}>
      <ViewerMembershipLayout queryRef={queryRef} />
    </Suspense>
  );
}

export default function () {
  return (
    <IAMRelayProvider>
      <MembershipLayoutLoader />
    </IAMRelayProvider>
  );
}
