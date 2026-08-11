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

import { useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalFrameworkList_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalFrameworkList_compliancePortalFragment.graphql";
import type { CompliancePortalFrameworkList_compliancePortalRefetchQuery } from "#/__generated__/core/CompliancePortalFrameworkList_compliancePortalRefetchQuery.graphql";

import { CompliancePortalFrameworkListItem } from "./CompliancePortalFrameworkListItem";

const compliancePortalFragment = graphql`
  fragment CompliancePortalFrameworkList_compliancePortalFragment on CompliancePortal
  @refetchable(queryName: "CompliancePortalFrameworkList_compliancePortalRefetchQuery")
  @argumentDefinitions(
    first: { type: Int, defaultValue: 100 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: ComplianceFrameworkOrder, defaultValue: { field: RANK, direction: ASC } }
  ) {
    ...CompliancePortalFrameworkListItem_compliancePortal
    complianceFrameworks(first: $first, after: $after, orderBy: $order)
    @connection(key: "CompliancePortalFrameworkList_complianceFrameworks", filters: ["orderBy"]) {
      edges {
        node {
          id
          framework {
            id
          }
          ...CompliancePortalFrameworkListItem_complianceFramework
        }
      }
    }
  }
`;

export interface CompliancePortalFrameworkListProps {
  compliancePortalRef: CompliancePortalFrameworkList_compliancePortalFragment$key;
}

export function CompliancePortalFrameworkList(props: CompliancePortalFrameworkListProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const [, startTransition] = useTransition();

  const [compliancePortal, refetch] = useRefetchableFragment<
    CompliancePortalFrameworkList_compliancePortalRefetchQuery,
    CompliancePortalFrameworkList_compliancePortalFragment$key
  >(compliancePortalFragment, props.compliancePortalRef);

  const edges = compliancePortal.complianceFrameworks.edges;

  const handleRefetch = () => {
    startTransition(() => {
      refetch({}, { fetchPolicy: "store-and-network" });
    });
  };

  if (edges.length === 0) {
    return (
      <p className="text-sm text-txt-secondary">
        {t("frameworkList.empty")}
      </p>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
      {edges.map(edge => (
        <CompliancePortalFrameworkListItem
          key={edge.node.framework.id}
          complianceFrameworkKey={edge.node}
          compliancePortalKey={compliancePortal}
          onRefetch={handleRefetch}
        />
      ))}
    </div>
  );
}
