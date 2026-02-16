import { useTranslate } from "@probo/i18n";
import { Layout, Skeleton } from "@probo/ui";
import { Suspense } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import type { ViewerMembershipLayoutQuery } from "#/__generated__/iam/ViewerMembershipLayoutQuery.graphql";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";
import { CurrentUser } from "#/providers/CurrentUser";

import { MembershipsDropdown } from "./_components/MembershipsDropdown";
import { Sidebar } from "./_components/Sidebar";
import { ViewerMembershipDropdown } from "./_components/ViewerMembershipDropdown";

export const viewerMembershipLayoutQuery = graphql`
  query ViewerMembershipLayoutQuery(
    $organizationId: ID!
    $hideSidebar: Boolean!
  ) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...MembershipsDropdown_organizationFragment
        ...ViewerMembershipDropdownFragment
        ...SidebarFragment @skip(if: $hideSidebar)
        viewer @required(action: THROW) {
          fullName
          membership @required(action: THROW) {
            role
          }
        }
      }
    }
    viewer @required(action: THROW) {
      email
    }
  }
`;

export function ViewerMembershipLayout(props: {
  hideSidebar?: boolean;
  queryRef: PreloadedQuery<ViewerMembershipLayoutQuery>;
}) {
  const { hideSidebar = false, queryRef } = props;

  const { __ } = useTranslate();

  const { organization, viewer }
    = usePreloadedQuery<ViewerMembershipLayoutQuery>(
      viewerMembershipLayoutQuery,
      queryRef,
    );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  return (
    <Layout
      headerLeading={(
        <MembershipsDropdown organizationFKey={organization} />
      )}
      headerTrailing={(
        <Suspense fallback={<Skeleton className="w-32 h-8" />}>
          <ViewerMembershipDropdown fKey={organization} />
        </Suspense>
      )}
      sidebar={!hideSidebar && <Sidebar fKey={organization} />}
    >
      <CoreRelayProvider>
        <CurrentUser
          value={{
            email: viewer.email,
            fullName: organization.viewer.fullName,
            role: organization.viewer.membership.role,
          }}
        >
          <Outlet context={organization.viewer.membership.role} />
        </CurrentUser>
      </CoreRelayProvider>
    </Layout>
  );
}
