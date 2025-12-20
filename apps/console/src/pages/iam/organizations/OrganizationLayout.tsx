import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { organizationLayoutQuery } from "./OrganizationLayoutQuery";
import { Link, Outlet } from "react-router";
import { Badge, Button, IconPeopleAdd, Layout, Skeleton } from "@probo/ui";
import { Sidebar } from "./_components/Sidebar";
import { OrganizationDropdown } from "./_components/OrganizationDropdown";
import type { OrganizationLayoutQuery } from "./__generated__/OrganizationLayoutQuery.graphql";
import { PermissionsProvider } from "/providers/PermissionsProvider";
import { SessionDropdown } from "./_components/SessionDropdown";
import { Suspense } from "react";
import { useTranslate } from "@probo/i18n";
import { CoreRelayProvider } from "/providers/CoreRelayProvider";

interface OrganizationLayoutProps {
  queryRef: PreloadedQuery<OrganizationLayoutQuery>;
}

export default function OrganizationLayout(props: OrganizationLayoutProps) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization, viewer } = usePreloadedQuery<OrganizationLayoutQuery>(
    organizationLayoutQuery,
    queryRef,
  );

  return (
    <PermissionsProvider>
      <Layout
        header={
          <>
            <div className="mr-auto">
              <OrganizationDropdown
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
              <SessionDropdown fKey={viewer} />
            </Suspense>
          </>
        }
        sidebar={<Sidebar />}
      >
        <CoreRelayProvider>
          <Outlet />
        </CoreRelayProvider>
      </Layout>
    </PermissionsProvider>
  );
}
