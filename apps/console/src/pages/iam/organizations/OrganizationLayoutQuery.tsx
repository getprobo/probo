import { lazy } from "@probo/react-lazy";
import { useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import { Suspense, useEffect } from "react";
import { Skeleton } from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { OrganizationLayoutQuery } from "./__generated__/OrganizationLayoutQuery.graphql";

const Layout = lazy(() => import("./OrganizationLayout"));

export const organizationLayoutQuery = graphql`
  query OrganizationLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      ... on Organization {
        ...OrganizationDropdownFragment
      }
    }
    viewer {
      pendingInvitations {
        totalCount
      }
    }
  }
`;

export function OrganizationLayoutQuery() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<OrganizationLayoutQuery>(
    organizationLayoutQuery,
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
