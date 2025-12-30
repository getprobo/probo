import { graphql, usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { Link, Outlet } from "react-router";
import { Badge, Button, IconPeopleAdd, Layout, Skeleton } from "@probo/ui";
import { Sidebar } from "./_components/Sidebar";
import { MembershipsDropdown } from "./MembershipsDropdown";
import type { MembershipLayoutQuery } from "/__generated__/iam/MembershipLayoutQuery.graphql";
import { PermissionsProvider } from "/providers/PermissionsProvider";
import { SessionDropdown } from "./_components/SessionDropdown";
import { Suspense } from "react";
import { useTranslate } from "@probo/i18n";
import { CoreRelayProvider } from "/providers/CoreRelayProvider";

export const membershipLayoutQuery = graphql`
  query MembershipLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      ... on Organization {
        ...MembershipsDropdown_organizationFragment
        ...SessionDropdownFragment
      }
    }
    viewer @required(action: THROW) {
      ...SidebarFragment @arguments(organizationId: $organizationId)
      ...MembershipsDropdown_viewerFragment
      pendingInvitations @required(action: THROW) {
        totalCount @required(action: THROW)
      }
    }
  }
`;

export function MembershipLayout(props: {
  queryRef: PreloadedQuery<MembershipLayoutQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization, viewer } = usePreloadedQuery<MembershipLayoutQuery>(
    membershipLayoutQuery,
    queryRef,
  );

  return (
    <PermissionsProvider>
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
        sidebar={<Sidebar fKey={viewer} />}
      >
        <CoreRelayProvider>
          <Outlet />
        </CoreRelayProvider>
      </Layout>
    </PermissionsProvider>
  );
}
