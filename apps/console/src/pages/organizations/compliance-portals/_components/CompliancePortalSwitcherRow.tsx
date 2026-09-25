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

import type { ReactNode } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePortalSwitcherRowQuery } from "#/__generated__/core/CompliancePortalSwitcherRowQuery.graphql";
import {
  navPanelSwitcher,
  NavPanelSwitcher,
  NavPanelSwitcherValueSkeleton,
} from "#/pages/organizations/_components/NavPanelSwitcher";

import { CompliancePortalOpenLink } from "./CompliancePortalOpenLink";
import { CompliancePortalSwitcherValue } from "./CompliancePortalSwitcherValue";

export const compliancePortalSwitcherRowQuery = graphql`
  query CompliancePortalSwitcherRowQuery($compliancePortalId: ID!) {
    node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalSwitcherValue_compliancePortal
        ...CompliancePortalOpenLink_compliancePortal
      }
    }
  }
`;

export interface CompliancePortalSwitcherRowProps {
  queryRef: PreloadedQuery<CompliancePortalSwitcherRowQuery>;
  label: string;
  onOpenChange: (open: boolean) => void;
  children?: ReactNode;
}

export function CompliancePortalSwitcherRow({
  queryRef,
  label,
  onOpenChange,
  children,
}: CompliancePortalSwitcherRowProps) {
  const slots = navPanelSwitcher();
  const data = usePreloadedQuery<CompliancePortalSwitcherRowQuery>(
    compliancePortalSwitcherRowQuery,
    queryRef,
  );
  const portal = data.node?.__typename === "CompliancePortal" ? data.node : null;

  return (
    <>
      <div className={slots.rowTrigger()}>
        <NavPanelSwitcher
          active={false}
          label={label}
          onOpenChange={onOpenChange}
          value={portal != null
            ? <CompliancePortalSwitcherValue compliancePortalKey={portal} />
            : <NavPanelSwitcherValueSkeleton />}
        >
          {children}
        </NavPanelSwitcher>
      </div>
      {portal != null && (
        <CompliancePortalOpenLink compliancePortalKey={portal} />
      )}
    </>
  );
}
