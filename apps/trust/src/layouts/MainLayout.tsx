import { useFavicon, useSystemTheme } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Logo, TabLink, Tabs } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import { NDADialog } from "#/components/NDADialog";
import { OrganizationSidebar } from "#/components/OrganizationSidebar";
import { TrustCenterProvider } from "#/providers/TrustCenterProvider";
import { Viewer } from "#/providers/Viewer";
import type { TrustGraphCurrentQuery } from "#/queries/__generated__/TrustGraphCurrentQuery.graphql";
import { currentTrustGraphQuery } from "#/queries/TrustGraph";

type Props = {
  queryRef: PreloadedQuery<TrustGraphCurrentQuery>;
};

export function MainLayout(props: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(currentTrustGraphQuery, props.queryRef);
  const trustCenter = data.currentTrustCenter;

  const theme = useSystemTheme();

  useFavicon(theme === "dark" ? (trustCenter?.darkLogoFileUrl ?? trustCenter?.logoFileUrl) : trustCenter?.logoFileUrl);

  if (!trustCenter) {
    return null;
  }

  const showNDADialog
    = trustCenter.isViewerMember
      && !trustCenter.hasAcceptedNonDisclosureAgreement
      && trustCenter.ndaFileUrl;
  return (
    <Viewer value={data.viewer}>
      <TrustCenterProvider trustCenter={trustCenter}>
        {showNDADialog && (
          <NDADialog
            organizationName={trustCenter.organization.name}
            url={trustCenter.ndaFileUrl}
            fileName={trustCenter.ndaFileName}
          />
        )}
        <div className="grid grid-cols-1 max-w-[1280px] mx-4 pt-6 gap-4 lg:mx-auto lg:gap-10 lg:pt-20 lg:grid-cols-[400px_1fr] lg:items-start ">
          <OrganizationSidebar trustCenter={trustCenter} />
          <main>
            <Tabs className="mb-8">
              <TabLink to="/overview">{__("Overview")}</TabLink>
              <TabLink to="/documents">{__("Documents")}</TabLink>
              {trustCenter.vendorInfo.totalCount > 0
                && <TabLink to="/subprocessors">{__("Subprocessors")}</TabLink>}
            </Tabs>
            <Outlet context={{ trustCenter }} />
          </main>
        </div>

        <a
          href="https://www.getprobo.com/"
          className="flex gap-2 text-sm font-medium text-txt-tertiary items-center w-max mx-auto my-10"
        >
          {__("Powered by")}
          {" "}
          <Logo withPicto className="h-6" />
        </a>
      </TrustCenterProvider>
    </Viewer>
  );
}
