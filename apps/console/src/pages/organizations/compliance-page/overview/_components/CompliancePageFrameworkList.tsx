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

import { useTranslate } from "@probo/i18n";
import { useTransition } from "react";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageFrameworkList_compliancePageFragment$key } from "#/__generated__/core/CompliancePageFrameworkList_compliancePageFragment.graphql";
import type { CompliancePageFrameworkList_compliancePageRefetchQuery } from "#/__generated__/core/CompliancePageFrameworkList_compliancePageRefetchQuery.graphql";

import { CompliancePageFrameworkListItem } from "./CompliancePageFrameworkListItem";

const compliancePageFragment = graphql`
  fragment CompliancePageFrameworkList_compliancePageFragment on TrustCenter
  @refetchable(queryName: "CompliancePageFrameworkList_compliancePageRefetchQuery")
  @argumentDefinitions(
    first: { type: Int, defaultValue: 100 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: ComplianceFrameworkOrder, defaultValue: { field: RANK, direction: ASC } }
  ) {
    ...CompliancePageFrameworkListItem_compliancePage
    complianceFrameworks(first: $first, after: $after, orderBy: $order)
    @connection(key: "CompliancePageFrameworkList_complianceFrameworks", filters: ["orderBy"]) {
      edges {
        node {
          id
          ...CompliancePageFrameworkListItem_complianceFramework
        }
      }
    }
  }
`;

export interface CompliancePageFrameworkListProps {
  compliancePageRef: CompliancePageFrameworkList_compliancePageFragment$key;
}

export function CompliancePageFrameworkList(props: CompliancePageFrameworkListProps) {
  const { __ } = useTranslate();
  const [, startTransition] = useTransition();

  const [compliancePage, refetch] = useRefetchableFragment<
    CompliancePageFrameworkList_compliancePageRefetchQuery,
    CompliancePageFrameworkList_compliancePageFragment$key
  >(compliancePageFragment, props.compliancePageRef);

  const edges = compliancePage.complianceFrameworks.edges;

  const handleRefetch = () => {
    startTransition(() => {
      refetch({}, { fetchPolicy: "store-and-network" });
    });
  };

  if (edges.length === 0) {
    return (
      <p className="text-sm text-txt-secondary">
        {__("No frameworks available")}
      </p>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
      {edges.map(edge => (
        <CompliancePageFrameworkListItem
          key={edge.node.id}
          complianceFrameworkKey={edge.node}
          compliancePageKey={compliancePage}
          onRefetch={handleRefetch}
        />
      ))}
    </div>
  );
}
