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
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery, useQueryLoader } from "react-relay";
import { useLocation, useParams } from "react-router";

import type { CompliancePortalNavItemsQuery } from "#/__generated__/core/CompliancePortalNavItemsQuery.graphql";
import type { CompliancePortalSwitcherRowQuery } from "#/__generated__/core/CompliancePortalSwitcherRowQuery.graphql";
import type { CompliancePortalNavPanel_organization$key } from "#/__generated__/iam/CompliancePortalNavPanel_organization.graphql";
import type { CompliancePortalNavPanelQuery } from "#/__generated__/iam/CompliancePortalNavPanelQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  CompliancePortalNavItems,
  compliancePortalNavItemsQuery,
} from "#/pages/organizations/compliance-portals/_components/CompliancePortalNavItems";
import { compliancePortalSwitcherRowQuery } from "#/pages/organizations/compliance-portals/_components/CompliancePortalSwitcherRow";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { NavPanelQuery } from "./NavPanelQuery";
import { navPanel } from "./variants";

const CompliancePortalSwitcher = lazy(async () => {
  const { CompliancePortalSwitcher: Component } = await import(
    "#/pages/organizations/compliance-portals/_components/CompliancePortalSwitcher"
  );
  return { default: Component };
});

const compliancePortalNavPanelQuery = graphql`
  query CompliancePortalNavPanelQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...CompliancePortalNavPanel_organization
      }
    }
  }
`;

const compliancePortalNavPanelFragment = graphql`
  fragment CompliancePortalNavPanel_organization on Organization {
    canGetCompliancePortal: permission(action: "compliance-portal:portal:get")
  }
`;

export function CompliancePortalNavPanel() {
  return (
    <NavPanelQuery<CompliancePortalNavPanelQuery> query={compliancePortalNavPanelQuery}>
      {queryRef => <CompliancePortalNavPanelInner queryRef={queryRef} />}
    </NavPanelQuery>
  );
}

interface CompliancePortalNavPanelInnerProps {
  queryRef: PreloadedQuery<CompliancePortalNavPanelQuery>;
}

function CompliancePortalNavPanelInner({ queryRef }: CompliancePortalNavPanelInnerProps) {
  const data = usePreloadedQuery<CompliancePortalNavPanelQuery>(
    compliancePortalNavPanelQuery,
    queryRef,
  );
  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }
  const organization = useFragment<CompliancePortalNavPanel_organization$key>(
    compliancePortalNavPanelFragment,
    data.organization,
  );

  if (!organization.canGetCompliancePortal) {
    return null;
  }

  return (
    <CoreRelayProvider>
      <CompliancePortalNavSection />
    </CoreRelayProvider>
  );
}

function CompliancePortalNavSection() {
  const organizationId = useOrganizationId();
  const { pathname } = useLocation();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const [rowQueryRef, loadRowQuery] = useQueryLoader<CompliancePortalSwitcherRowQuery>(
    compliancePortalSwitcherRowQuery,
  );
  const [itemsQueryRef, loadItemsQuery] = useQueryLoader<CompliancePortalNavItemsQuery>(
    compliancePortalNavItemsQuery,
  );
  const slots = navPanel();
  const prefix = `/organizations/${organizationId}/compliance-portals/`;
  const isNew = pathname === `${prefix}new`;
  const hasPortal = compliancePortalId != null && !isNew;
  const fallback = <span className={slots.groupFallback()} aria-hidden />;

  useEffect(() => {
    if (!hasPortal || compliancePortalId == null) {
      return;
    }
    startTransition(() => {
      loadRowQuery(
        { compliancePortalId },
        { fetchPolicy: "store-or-network" },
      );
      loadItemsQuery(
        { compliancePortalId },
        { fetchPolicy: "store-or-network" },
      );
    });
  }, [compliancePortalId, hasPortal, loadItemsQuery, loadRowQuery]);

  return (
    <>
      <Suspense fallback={fallback}>
        <CompliancePortalSwitcher queryRef={rowQueryRef ?? null} />
      </Suspense>
      {itemsQueryRef != null && (
        <Suspense fallback={null}>
          <CompliancePortalNavItems queryRef={itemsQueryRef} />
        </Suspense>
      )}
    </>
  );
}
