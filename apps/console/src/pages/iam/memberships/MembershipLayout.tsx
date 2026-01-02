import { graphql, usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { Link, Outlet } from "react-router";
import { Badge, Button, IconPeopleAdd, Layout, Skeleton } from "@probo/ui";
import { Sidebar } from "./_components/Sidebar";
import { MembershipsDropdown } from "./MembershipsDropdown";
import type { MembershipLayoutQuery } from "/__generated__/iam/MembershipLayoutQuery.graphql";
import { SessionDropdown } from "./_components/SessionDropdown";
import { Suspense } from "react";
import { useTranslate } from "@probo/i18n";
import { CoreRelayProvider } from "/providers/CoreRelayProvider";
import { CurrentUser } from "/providers/CurrentUser";

export const membershipLayoutQuery = graphql`
  query MembershipLayoutQuery($organizationId: ID!, $hideSidebar: Boolean!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...MembershipsDropdown_organizationFragment
        ...SessionDropdownFragment
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

export function MembershipLayout(props: {
  hideSidebar?: boolean;
  queryRef: PreloadedQuery<MembershipLayoutQuery>;
}) {
  const { hideSidebar = false, queryRef } = props;

  const { __ } = useTranslate();

  const { organization, viewer } = usePreloadedQuery<MembershipLayoutQuery>(
    membershipLayoutQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  return (
    <Layout
      header={
        <>
          <div className="mr-auto">
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
          </div>
          <Suspense fallback={<Skeleton className="w-32 h-8" />}>
            <SessionDropdown fKey={organization} />
          </Suspense>
        </>
      }
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
