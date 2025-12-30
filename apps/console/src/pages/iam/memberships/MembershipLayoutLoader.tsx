import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";
import { Skeleton } from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { MembershipLayoutQuery } from "/__generated__/iam/MembershipLayoutQuery.graphql";
import { MembershipLayout, membershipLayoutQuery } from "./MembershipLayout";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

function MembershipLayoutLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<MembershipLayoutQuery>(
    membershipLayoutQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <Skeleton className="w-full h-screen" />;
  }

  return (
    <Suspense fallback={<Skeleton className="w-full h-screen" />}>
      <MembershipLayout queryRef={queryRef} />
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
