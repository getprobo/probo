import { lazy } from "@probo/react-lazy";
import { useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import { Suspense, useEffect } from "react";
import { Skeleton } from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { MembershipLayoutQuery } from "./__generated__/MembershipLayoutQuery.graphql";

const Layout = lazy(() => import("./MembershipLayout"));

export const membershipLayoutQuery = graphql`
  query MembershipLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      ... on Organization {
        ...MembershipsDropdown_organizationFragment
      }
    }
    viewer @required(action: THROW) {
      ...MembershipsDropdown_viewerFragment
      pendingInvitations @required(action: THROW) {
        totalCount @required(action: THROW)
      }
      ...SessionDropdownFragment @arguments(organizationId: $organizationId)
    }
  }
`;

export function MembershipLayoutQuery() {
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
      <Layout queryRef={queryRef} />
    </Suspense>
  );
}
