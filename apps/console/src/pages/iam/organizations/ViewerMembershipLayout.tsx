import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconPeopleAdd, Layout, Skeleton } from "@probo/ui";
import { Suspense } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link, Outlet } from "react-router";

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
        viewerMembership @required(action: THROW) {
          role
          profile @required(action: THROW) {
            fullName
          }
        }
      }
    }
    viewer @required(action: THROW) {
      email
      ...MembershipsDropdown_viewerFragment
      pendingInvitations @required(action: THROW) {
        totalCount @required(action: THROW)
      }
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
        <>
          <MembershipsDropdown
            organizationFKey={organization}
            viewerFKey={viewer}
          />
          {viewer.pendingInvitations.totalCount > 0 && (
            <Link to="/" className="relative" title={__("Invitations")}>
              <Button variant="tertiary" icon={IconPeopleAdd} />
              <Badge
                variant="info"
                size="sm"
                className="absolute -top-1 -right-1 min-w-[20px] h-5 flex items-center justify-center"
              >
                {viewer.pendingInvitations.totalCount}
              </Badge>
            </Link>
          )}
        </>
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
            fullName: organization.viewerMembership.profile.fullName,
            role: organization.viewerMembership.role,
          }}
        >
          <Outlet context={organization.viewerMembership.role} />
        </CurrentUser>
      </CoreRelayProvider>
    </Layout>
  );
}
