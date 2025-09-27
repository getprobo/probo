import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import type { TrustGraphQuery } from "/queries/__generated__/TrustGraphQuery.graphql.ts";
import { trustGraphQuery } from "/queries/TrustGraph.ts";
import { Tabs, TabLink, Logo } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { OrganizationSidebar } from "/components/OrganizationSidebar";
import { Outlet } from "react-router";
import { NDADialog } from "/components/NDADialog";
import { useState } from "react";
import { AuthProvider } from "/providers/AuthProvider";
import { TrustCenterProvider } from "/providers/TrustCenterProvider";

type Props = {
  queryRef: PreloadedQuery<TrustGraphQuery>;
};

export function MainLayout(props: Props) {
  const { __ } = useTranslate();
  const trustCenter = usePreloadedQuery(
    trustGraphQuery,
    props.queryRef
  ).trustCenterBySlug;

  if (!trustCenter) {
    return null;
  }

  const baseTabUrl = `/trust/${trustCenter.slug}`;
  const [showNDADialog] = useState(
    trustCenter.isUserAuthenticated &&
      !trustCenter.hasAcceptedNonDisclosureAgreement
  );

  return (
    <AuthProvider isAuthenticated={trustCenter.isUserAuthenticated}>
      <TrustCenterProvider trustCenter={trustCenter}>
        {showNDADialog && (
          <NDADialog
            name={trustCenter.organization.name}
            trustCenterId={trustCenter.id}
            url={trustCenter.ndaFileUrl}
            fileName={trustCenter.ndaFileName}
          />
        )}
        <div className="grid grid-cols-1 max-w-[1280px] mx-4 pt-6 gap-4 lg:mx-auto lg:gap-10 lg:pt-20 lg:grid-cols-[400px_1fr] lg:items-start ">
          <OrganizationSidebar trustCenter={trustCenter} />
          <main>
            <Tabs className="mb-8">
              <TabLink to={`${baseTabUrl}/overview`}>{__("Overview")}</TabLink>
              <TabLink to={`${baseTabUrl}/documents`}>
                {__("Documents")}
              </TabLink>
              <TabLink to={`${baseTabUrl}/subprocessors`}>
                {__("Subprocessors")}
              </TabLink>
            </Tabs>
            <Outlet context={{ trustCenter }} />
          </main>
        </div>

        <div className="flex gap-2 text-sm font-medium text-txt-tertiary items-center w-max mx-auto my-10">
          {__("Powered by")} <Logo withPicto className="h-6" />
        </div>
      </TrustCenterProvider>
    </AuthProvider>
  );
}
