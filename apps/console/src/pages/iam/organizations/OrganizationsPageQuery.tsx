import { lazy } from "@probo/react-lazy";
import { useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import { Suspense, useEffect } from "react";
import { CenteredLayoutSkeleton } from "@probo/ui";
import type { OrganizationsPageQuery } from "./__generated__/OrganizationsPageQuery.graphql";

const Page = lazy(() => import("./OrganizationsPage"));

export const organizationsPageQuery = graphql`
  query OrganizationsPageQuery {
    viewer @required(action: THROW) {
      memberships(first: 1000, orderBy: { direction: DESC, field: CREATED_AT })
        @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            ...MembershipCardFragment
            organization @required(action: THROW) {
              name
            }
          }
        }
      }
      pendingInvitations(
        first: 1000
        orderBy: { direction: DESC, field: CREATED_AT }
      ) @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            ...InvitationCardFragment
          }
        }
      }
    }
  }
`;

export function OrganizationsPageQuery() {
  const [queryRef, loadQuery] = useQueryLoader<OrganizationsPageQuery>(
    organizationsPageQuery,
  );

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
