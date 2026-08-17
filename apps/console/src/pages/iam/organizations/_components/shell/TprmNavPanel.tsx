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
import { graphql, useFragment, useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { ThirdPartySwitcherValueQuery } from "#/__generated__/core/ThirdPartySwitcherValueQuery.graphql";
import type { TprmNavPanel_organization$key } from "#/__generated__/iam/TprmNavPanel_organization.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navHref } from "#/pages/iam/organizations/_lib/navigation";
import { ThirdPartyNavItems } from "#/pages/organizations/third-parties/_components/ThirdPartyNavItems";
import { thirdPartySwitcherValueQuery } from "#/pages/organizations/third-parties/_components/ThirdPartySwitcherValue";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { NavPanelItem } from "./NavPanelItem";
import type { NavPanelBodyProps } from "./navPanels";
import { navPanel } from "./variants";

const ThirdPartySwitcher = lazy(async () => {
  const { ThirdPartySwitcher: Component } = await import(
    "#/pages/organizations/third-parties/_components/ThirdPartySwitcher"
  );
  return { default: Component };
});

const tprmNavPanelFragment = graphql`
  fragment TprmNavPanel_organization on Organization {
    canListThirdParties: permission(action: "core:thirdParty:list")
  }
`;

export function TprmNavPanel({ organizationKey, group }: NavPanelBodyProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const organization = useFragment<TprmNavPanel_organization$key>(
    tprmNavPanelFragment,
    organizationKey,
  );

  if (!organization.canListThirdParties) {
    return null;
  }

  return (
    <>
      <NavPanelItem
        label={t("nav.allThirdParties")}
        to={navHref(organizationId, group, "third-parties")}
        exact
      />
      <CoreRelayProvider>
        <ThirdPartyNavSection />
      </CoreRelayProvider>
    </>
  );
}

function ThirdPartyNavSection() {
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<ThirdPartySwitcherValueQuery>(
    thirdPartySwitcherValueQuery,
  );
  const slots = navPanel();
  const fallback = <span className={slots.groupFallback()} aria-hidden />;

  useEffect(() => {
    if (thirdPartyId == null) {
      return;
    }
    startTransition(() => {
      loadQuery(
        { thirdPartyId },
        { fetchPolicy: "store-or-network" },
      );
    });
  }, [loadQuery, thirdPartyId]);

  return (
    <>
      <Suspense fallback={fallback}>
        <ThirdPartySwitcher queryRef={queryRef ?? null} />
      </Suspense>
      <ThirdPartyNavItems />
    </>
  );
}
