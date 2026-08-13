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
import { useMemo, useState } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import type { OrganizationLayoutQuery } from "#/__generated__/iam/OrganizationLayoutQuery.graphql";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";
import { CurrentUser } from "#/providers/CurrentUser";

import { NavPanel } from "./_components/shell/NavPanel";
import { NavRail } from "./_components/shell/NavRail";
import { TopBar } from "./_components/shell/TopBar";
import { organizationLayout } from "./_components/shell/variants";

export const organizationLayoutQuery = graphql`
  query OrganizationLayoutQuery(
    $organizationId: ID!
    $hideNavigation: Boolean!
  ) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...TopBar_organization @include(if: $hideNavigation)
        ...NavRail_organization @skip(if: $hideNavigation)
        ...NavPanel_organization @skip(if: $hideNavigation)
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

export interface OrganizationLayoutProps {
  // The employee portal reuses this layout but has no products to navigate.
  // It falls back to the top bar, which is the only place its occupants can
  // switch organization or sign out.
  hideNavigation?: boolean;
  queryRef: PreloadedQuery<OrganizationLayoutQuery>;
}

/**
 * Chrome for the organization area: the rail, the panel of the active product,
 * and the page. The rail carries the organization and the account as its first
 * and last rows, so there is no top bar — except in the employee portal, which
 * has no rail to put them in.
 *
 * The frame owns the viewport — the page scrolls inside `content` rather than
 * on the document, which is what holds the chrome in place without any of it
 * being `position: fixed`.
 */
export function OrganizationLayout({ hideNavigation = false, queryRef }: OrganizationLayoutProps) {
  const { organization, viewer } = usePreloadedQuery<OrganizationLayoutQuery>(
    organizationLayoutQuery,
    queryRef,
  );

  // A v1 Drawer reports itself through this context so the page can make room
  // for it. The v1 Layout used to host it; this layout takes that over for the
  // pages that still render one.
  const [hasDrawer, setDrawer] = useState(false);
  const drawerContext = useMemo(() => ({ setDrawer }), []);

  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  const slots = organizationLayout({ hasDrawer });

  return (
    <LayoutContext value={drawerContext}>
      <div className={slots.root()}>
        {hideNavigation && <TopBar organizationKey={organization} />}
        <div className={slots.body()}>
          {!hideNavigation && <NavRail organizationKey={organization} />}
          {!hideNavigation && <NavPanel organizationKey={organization} />}
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
