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
import { ErrorBoundary } from "react-error-boundary";
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { useLocation, useParams } from "react-router";

import type { CookieBannerSwitcherValueQuery } from "#/__generated__/core/CookieBannerSwitcherValueQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CookieBannerNavItems } from "#/pages/organizations/cookie-banners/_components/CookieBannerNavItems";
import {
  cookieBannerFromSwitcherValueQuery,
  cookieBannerSwitcherValueQuery,
} from "#/pages/organizations/cookie-banners/_components/CookieBannerSwitcherValue";
import { cookieBannersBasePath } from "#/pages/organizations/cookie-banners/_lib/cookieBannerPaths";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import type { NavPanelBodyProps } from "./navPanels";
import { navPanel } from "./variants";

const CookieBannerSwitcher = lazy(async () => {
  const { CookieBannerSwitcher: Component } = await import(
    "#/pages/organizations/cookie-banners/_components/CookieBannerSwitcher"
  );
  return { default: Component };
});

export function CmpNavPanel(_: NavPanelBodyProps) {
  return (
    <CoreRelayProvider>
      <CookieBannerNavSection />
    </CoreRelayProvider>
  );
}

function CookieBannerNavSection() {
  const organizationId = useOrganizationId();
  const { pathname } = useLocation();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<CookieBannerSwitcherValueQuery>(
    cookieBannerSwitcherValueQuery,
  );
  const slots = navPanel();
  const isNew = pathname === `${cookieBannersBasePath(organizationId)}/new`;
  const fallback = <span className={slots.groupFallback()} aria-hidden />;
  const switcher = (
    <Suspense fallback={fallback}>
      <CookieBannerSwitcher banner={null} />
    </Suspense>
  );

  useEffect(() => {
    if (isNew) {
      return;
    }
    startTransition(() => {
      loadQuery(
        {
          organizationId,
          cookieBannerId: cookieBannerId ?? "",
          hasCookieBannerId: cookieBannerId != null,
        },
        { fetchPolicy: "store-or-network" },
      );
    });
  }, [cookieBannerId, isNew, loadQuery, organizationId]);

  if (isNew) {
    return switcher;
  }

  // loadQuery runs in an effect, so right after a banner-to-banner navigation
  // the ref still describes the previous banner. Rendering it would flash the
  // old name in the switcher and point the nav items at the old banner.
  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    && queryRef.variables.cookieBannerId === (cookieBannerId ?? "")
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return switcher;
  }

  return (
    <ErrorBoundary key={cookieBannerId ?? organizationId} fallbackRender={() => switcher}>
      <Suspense fallback={switcher}>
        <CookieBannerNavSelection queryRef={currentQueryRef} />
      </Suspense>
    </ErrorBoundary>
  );
}

interface CookieBannerNavSelectionProps {
  queryRef: PreloadedQuery<CookieBannerSwitcherValueQuery>;
}

function CookieBannerNavSelection({ queryRef }: CookieBannerNavSelectionProps) {
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<CookieBannerSwitcherValueQuery>(
    cookieBannerSwitcherValueQuery,
    queryRef,
  );
  const banner = cookieBannerFromSwitcherValueQuery(data, organizationId);
  const id = banner?.id ?? null;
  const tcf = banner?.capabilities.tcf ?? false;

  return (
    <>
      <CookieBannerSwitcher banner={banner} />
      {id != null && <CookieBannerNavItems cookieBannerId={id} tcf={tcf} />}
    </>
  );
}
