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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useLazyLoadQuery } from "react-relay";
import { useParams } from "react-router";

import type { CompliancePortalSwitcherValueQuery } from "#/__generated__/core/CompliancePortalSwitcherValueQuery.graphql";

import { compliancePortalSwitcher } from "./variants";

const compliancePortalSwitcherValueQuery = graphql`
  query CompliancePortalSwitcherValueQuery($compliancePortalId: ID!) {
    node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        entityName
      }
    }
  }
`;

export interface CompliancePortalSwitcherValueProps {
  fallback: string;
}

/**
 * The entity name of the portal the URL is on, or the default trigger label.
 *
 * Mounted only when `:compliancePortalId` is present. The create form uses
 * "New Portal" in the trigger itself.
 */
export function CompliancePortalSwitcherValue({ fallback }: CompliancePortalSwitcherValueProps) {
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const data = useLazyLoadQuery<CompliancePortalSwitcherValueQuery>(
    compliancePortalSwitcherValueQuery,
    { compliancePortalId: compliancePortalId ?? "" },
    { fetchPolicy: "store-or-network" },
  );
  const slots = compliancePortalSwitcher();
  if (data.node?.__typename !== "CompliancePortal") {
    return fallback;
  }
  return (
    <Text size={2} weight="medium" color="neutral" highContrast className={slots.itemName()}>
      {data.node.entityName}
    </Text>
  );
}

/**
 * A pulse bar the same height as the trigger name, used while the selected
 * portal query suspends. A plain "Select Portal" string inherited the page
 * font and flashed larger than the closed trigger.
 */
export function CompliancePortalSwitcherValueSkeleton() {
  const slots = compliancePortalSwitcher();

  return <span className={slots.valueSkeletonName()} aria-hidden />;
}
