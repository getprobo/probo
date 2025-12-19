import {
  RelayEnvironmentProvider,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { organizationLayoutQuery } from "./OrganizationLayoutQuery";
import { Outlet } from "react-router";
import { Layout, Skeleton } from "@probo/ui";
import { Sidebar } from "./_components/Sidebar";
import { OrganizationDropdown } from "./_components/OrganizationDropdown";
import { consoleEnvironment } from "/environments";
import type { OrganizationLayoutQuery } from "./__generated__/OrganizationLayoutQuery.graphql";
import { PermissionsProvider } from "/providers/PermissionsProvider";
import { SessionDropdown } from "./_components/SessionDropdown";
import { Suspense } from "react";

interface OrganizationLayoutProps {
  queryRef: PreloadedQuery<OrganizationLayoutQuery>;
}

export default function OrganizationLayout(props: OrganizationLayoutProps) {
  const { queryRef } = props;

  const data = usePreloadedQuery<OrganizationLayoutQuery>(
    organizationLayoutQuery,
    queryRef,
  );

  return (
    <PermissionsProvider>
      <Layout
        header={
          <>
            <div className="mr-auto">
              <OrganizationDropdown fKey={data.organization} />
            </div>
            <Suspense fallback={<Skeleton className="w-32 h-8" />}>
              <SessionDropdown fKey={data.viewer} />
            </Suspense>
          </>
        }
        sidebar={<Sidebar />}
      >
        <RelayEnvironmentProvider environment={consoleEnvironment}>
          <Outlet />
        </RelayEnvironmentProvider>
      </Layout>
    </PermissionsProvider>
  );
}
