import { lazy } from "@probo/react-lazy";
import { useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import type { OrganizationsQuery } from "./__generated__/OrganizationsQuery.graphql";
import { Suspense, useEffect } from "react";
import { CenteredLayoutSkeleton } from "@probo/ui";

const Page = lazy(() => import("./OrganizationsPage"));

export const organizationsQuery = graphql`
  query OrganizationsQuery {
    viewer @required(action: THROW) {
      memberships(
        first: 1000
        orderBy: { direction: DESC, field: CREATED_AT }
      ) {
        edges {
          node {
            id
            ...MembershipCardFragment
            organization {
              name
            }
          }
        }
      }
      pendingInvitations(
        first: 1000
        orderBy: { direction: DESC, field: CREATED_AT }
      ) {
        edges {
          node {
            id
            ...InvitationCardFragment
          }
        }
      }
    }
  }
`;

export function OrganizationsQuery() {
  const [queryRef, loadQuery] =
    useQueryLoader<OrganizationsQuery>(organizationsQuery);

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) {
    return <CenteredLayoutSkeleton />;
  }

  return (
    <Suspense fallback={<CenteredLayoutSkeleton />}>
      <Page queryRef={queryRef} />
    </Suspense>
  );
}
