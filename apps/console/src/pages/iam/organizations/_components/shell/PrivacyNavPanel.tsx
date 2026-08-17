// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { lazy } from "@probo/react-lazy";
import { startTransition, Suspense, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { useLocation } from "react-router";

import type { CookieBannerSwitcherValueQuery } from "#/__generated__/core/CookieBannerSwitcherValueQuery.graphql";
import type { PrivacyNavPanelQuery } from "#/__generated__/iam/PrivacyNavPanelQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navHref } from "#/pages/iam/organizations/_lib/navigation";
import { CookieBannerNavItems } from "#/pages/organizations/cookie-banners/_components/CookieBannerNavItems";
import { cookieBannerSwitcherValueQuery } from "#/pages/organizations/cookie-banners/_components/CookieBannerSwitcherValue";
import { cookieBannersBasePath } from "#/pages/organizations/cookie-banners/_lib/cookieBannerPaths";
import { useSelectedCookieBannerId } from "#/pages/organizations/cookie-banners/_lib/useSelectedCookieBannerId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { NavPanelGroup } from "./NavPanelGroup";
import { NavPanelItem } from "./NavPanelItem";
import { NavPanelQuery } from "./NavPanelQuery";
import type { NavPanelBodyProps } from "./navPanels";
import { navPanel } from "./variants";

const CookieBannerSwitcher = lazy(async () => {
  const { CookieBannerSwitcher: Component } = await import(
    "#/pages/organizations/cookie-banners/_components/CookieBannerSwitcher"
  );
  return { default: Component };
});

const privacyNavPanelQuery = graphql`
  query PrivacyNavPanelQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canListRightsRequests: permission(action: "core:rights-request:list")
        canListProcessingActivities: permission(action: "core:processing-activity:list")
        canListDataProtectionImpactAssessments: permission(action: "core:data-protection-impact-assessment:list")
        canListTransferImpactAssessments: permission(action: "core:transfer-impact-assessment:list")
        canListCookieBanners: permission(action: "core:cookie-banner:list")
      }
    }
  }
`;

export function PrivacyNavPanel({ group }: NavPanelBodyProps) {
  return (
    <NavPanelQuery<PrivacyNavPanelQuery> query={privacyNavPanelQuery}>
      {queryRef => <PrivacyNavPanelInner queryRef={queryRef} group={group} />}
    </NavPanelQuery>
  );
}

interface PrivacyNavPanelInnerProps extends NavPanelBodyProps {
  queryRef: PreloadedQuery<PrivacyNavPanelQuery>;
}

function PrivacyNavPanelInner({ queryRef, group }: PrivacyNavPanelInnerProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<PrivacyNavPanelQuery>(privacyNavPanelQuery, queryRef);
  const { organization } = data;
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  return (
    <>
      {organization.canListRightsRequests && (
        <NavPanelItem
          label={t("nav.rightsRequests")}
          to={navHref(organizationId, group, "rights-requests")}
        />
      )}
      {organization.canListProcessingActivities && (
        <NavPanelItem
          label={t("nav.processingActivities")}
          to={navHref(organizationId, group, "processing-activities")}
        />
      )}
      {organization.canListDataProtectionImpactAssessments && (
        <NavPanelItem
          label={t("nav.dataProtectionImpactAssessments")}
          to={navHref(organizationId, group, "dpias")}
        />
      )}
      {organization.canListTransferImpactAssessments && (
        <NavPanelItem
          label={t("nav.transferImpactAssessments")}
          to={navHref(organizationId, group, "tias")}
        />
      )}
      {organization.canListCookieBanners && (
        <NavPanelGroup label={t("nav.cookieBanners")}>
          <CoreRelayProvider>
            <CookieBannerNavSection />
          </CoreRelayProvider>
        </NavPanelGroup>
      )}
    </>
  );
}

function CookieBannerNavSection() {
  const organizationId = useOrganizationId();
  const { pathname } = useLocation();
  const selectedId = useSelectedCookieBannerId();
  const [queryRef, loadQuery] = useQueryLoader<CookieBannerSwitcherValueQuery>(
    cookieBannerSwitcherValueQuery,
  );
  const slots = navPanel();
  const isNew = pathname === `${cookieBannersBasePath(organizationId)}/new`;
  const fallback = <span className={slots.groupFallback()} aria-hidden />;

  useEffect(() => {
    if (isNew) {
      return;
    }
    startTransition(() => {
      loadQuery(
        {
          organizationId,
          cookieBannerId: selectedId ?? "",
          hasCookieBannerId: selectedId != null,
        },
        { fetchPolicy: "store-or-network" },
      );
    });
  }, [isNew, loadQuery, organizationId, selectedId]);

  if (isNew) {
    return (
      <Suspense fallback={fallback}>
        <CookieBannerSwitcher queryRef={null} selectedId={selectedId} />
      </Suspense>
    );
  }

  // loadQuery runs in an effect, so right after a banner-to-banner navigation
  // the ref still describes the previous banner. Rendering it would flash the
  // old name in the switcher and point the nav items at the old banner.
  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    && queryRef.variables.cookieBannerId === (selectedId ?? "")
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return fallback;
  }

  return (
    <>
      <Suspense fallback={fallback}>
        <CookieBannerSwitcher queryRef={currentQueryRef} selectedId={selectedId} />
      </Suspense>
      <Suspense fallback={null}>
        <CookieBannerNavItems queryRef={currentQueryRef} />
      </Suspense>
    </>
  );
}
