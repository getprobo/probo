import { lazy } from "@probo/react-lazy";
import { useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import { Suspense, useEffect } from "react";
import { CenteredLayoutSkeleton } from "@probo/ui";
import type { MembershipsPageQuery } from "./__generated__/MembershipsPageQuery.graphql";

const Page = lazy(() => import("./MembershipsPage"));

export const membershipsPageQuery = graphql`
  query MembershipsPageQuery {
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

export function MembershipsPageQuery() {
  const [queryRef, loadQuery] =
    useQueryLoader<MembershipsPageQuery>(membershipsPageQuery);

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
