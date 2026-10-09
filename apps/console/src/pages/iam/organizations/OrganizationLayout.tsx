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

import { LayoutContext } from "@probo/ui";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Suspense, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { Outlet } from "react-router";

import type { OrganizationLayoutQuery } from "#/__generated__/iam/OrganizationLayoutQuery.graphql";
import type { ViewerMembershipMenuEmployeePortalQuery } from "#/__generated__/iam/ViewerMembershipMenuEmployeePortalQuery.graphql";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";
import { CurrentUser } from "#/providers/CurrentUser";

import { NavPanel } from "./_components/shell/NavPanel";
import { NavRail } from "./_components/shell/NavRail";
import { NavSpotlight } from "./_components/shell/NavSpotlight";
import { organizationLayout } from "./_components/shell/variants";
import { viewerMembershipMenuEmployeePortalQuery } from "./_components/shell/ViewerMembershipMenu";
import { OrganizationLayoutSkeleton } from "./OrganizationLayoutSkeleton";

export const organizationLayoutQuery = graphql`
  query OrganizationLayoutQuery($organizationId: ID!) {
    slackbotAvailable
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canListEmployeePortals: permission(action: "employee-portal:portal:list")
        ...NavRail_organization
        ...NavPanel_organization
        ...NavSpotlight_organization
        viewer @required(action: THROW) {
          fullName
          membership @required(action: THROW) {
            role
          }
        }
      }
    }
    viewer @required(action: THROW) {
      email
    }
  }
`;

interface OrganizationLayoutProps {
  queryRef: PreloadedQuery<OrganizationLayoutQuery>;
  employeePortalQueryRef: PreloadedQuery<ViewerMembershipMenuEmployeePortalQuery> | null;
}

function EmployeePortalQueryProbe({
  queryRef: portalQueryRef,
  onReady,
}: {
  queryRef: PreloadedQuery<ViewerMembershipMenuEmployeePortalQuery>;
  onReady: () => void;
}) {
  usePreloadedQuery<ViewerMembershipMenuEmployeePortalQuery>(
    viewerMembershipMenuEmployeePortalQuery,
    portalQueryRef,
  );
  useEffect(() => {
    onReady();
  }, [onReady]);
  return null;
}

export function OrganizationLayoutGate({ queryRef }: { queryRef: PreloadedQuery<OrganizationLayoutQuery> }) {
  const data = usePreloadedQuery<OrganizationLayoutQuery>(organizationLayoutQuery, queryRef);
  const { organization } = data;
  const { organizationId } = queryRef.variables;
  const canListEmployeePortals = organization.__typename === "Organization"
    && organization.canListEmployeePortals;
  const [portalQueryRef, loadPortalQuery] = useQueryLoader<ViewerMembershipMenuEmployeePortalQuery>(
    viewerMembershipMenuEmployeePortalQuery,
  );
  const [shellVisible, setShellVisible] = useState(false);
  const [blockedFetchKey, setBlockedFetchKey] = useState<string | number | null>(null);
  const retriedPortal = useRef(false);
  const [seenOrganizationId, setSeenOrganizationId] = useState(organizationId);
  const showShell = useCallback(() => {
    setShellVisible(true);
  }, []);
  if (seenOrganizationId !== organizationId) {
    setSeenOrganizationId(organizationId);
    setShellVisible(false);
    setBlockedFetchKey(null);
  }

  useEffect(() => {
    if (!canListEmployeePortals) {
      return;
    }
    retriedPortal.current = false;
    loadPortalQuery({ organizationId }, { fetchPolicy: "store-or-network" });
  }, [canListEmployeePortals, loadPortalQuery, organizationId]);

  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  if (!canListEmployeePortals) {
    return (
      <OrganizationLayout
        queryRef={queryRef}
        employeePortalQueryRef={null}
      />
    );
  }

  const readyPortalQueryRef = portalQueryRef != null
    && portalQueryRef.variables.organizationId === organizationId
    ? portalQueryRef
    : null;

  const portalForMenu = readyPortalQueryRef != null
    && readyPortalQueryRef.fetchKey !== blockedFetchKey
    ? readyPortalQueryRef
    : null;

  if (!shellVisible) {
    return (
      <>
        <OrganizationLayoutSkeleton />
        {readyPortalQueryRef != null && (
          <ErrorBoundary
            onError={() => {
              setShellVisible(true);
              setBlockedFetchKey(readyPortalQueryRef.fetchKey);
              if (retriedPortal.current) {
                return;
              }
              retriedPortal.current = true;
              // A failed preloaded query stays failed until loadQuery returns a new ref.
              loadPortalQuery({ organizationId }, { fetchPolicy: "network-only" });
            }}
            fallback={null}
          >
            <Suspense fallback={null}>
              <EmployeePortalQueryProbe
                queryRef={readyPortalQueryRef}
                onReady={showShell}
              />
            </Suspense>
          </ErrorBoundary>
        )}
      </>
    );
  }

  return (
    <OrganizationLayout
      queryRef={queryRef}
      employeePortalQueryRef={portalForMenu}
    />
  );
}

export function OrganizationLayout({
  queryRef,
  employeePortalQueryRef,
}: OrganizationLayoutProps) {
  const { organization, viewer, slackbotAvailable } = usePreloadedQuery<OrganizationLayoutQuery>(
    organizationLayoutQuery,
    queryRef,
  );

  const [hasDrawer, setDrawer] = useState(false);
  const drawerContext = useMemo(() => ({ setDrawer }), []);

  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  const slots = organizationLayout({ hasDrawer });

  return (
    <LayoutContext value={drawerContext}>
      <div className={slots.root()}>
        <div className={slots.body()}>
          <NavSpotlight
            organizationKey={organization}
            slackbotAvailable={slackbotAvailable}
          />
          <NavRail
            organizationKey={organization}
            slackbotAvailable={slackbotAvailable}
            employeePortalQueryRef={employeePortalQueryRef}
          />
          <NavPanel
            organizationKey={organization}
            slackbotAvailable={slackbotAvailable}
          />
          <main className={slots.content()}>
            <div className={slots.contentInner()}>
              <CoreRelayProvider>
                <CurrentUser
                  value={{
                    email: viewer.email,
                    fullName: organization.viewer.fullName,
                    role: organization.viewer.membership.role,
                  }}
                >
                  <Outlet context={organization.viewer.membership.role} />
                </CurrentUser>
              </CoreRelayProvider>
            </div>
          </main>
        </div>
      </div>
    </LayoutContext>
  );
}
