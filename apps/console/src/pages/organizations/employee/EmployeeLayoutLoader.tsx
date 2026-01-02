import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";
import { Skeleton } from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";
import {
  MembershipLayout,
  membershipLayoutQuery,
} from "/pages/iam/memberships/MembershipLayout";
import type { MembershipLayoutQuery } from "/__generated__/iam/MembershipLayoutQuery.graphql";

function EmployeeLayoutLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<MembershipLayoutQuery>(
    membershipLayoutQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
      hideSidebar: true,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <Skeleton className="w-full h-screen" />;
  }

  return (
    <Suspense fallback={<Skeleton className="w-full h-screen" />}>
      <MembershipLayout queryRef={queryRef} hideSidebar />
    </Suspense>
  );
}

export default function () {
  return (
    <IAMRelayProvider>
      <EmployeeLayoutLoader />
    </IAMRelayProvider>
  );
}
