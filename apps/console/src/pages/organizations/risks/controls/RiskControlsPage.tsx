// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Badge, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { RiskControlsPage_risk$key } from "#/__generated__/core/RiskControlsPage_risk.graphql";
import type { RiskControlsPageQuery } from "#/__generated__/core/RiskControlsPageQuery.graphql";
import type { RiskControlsPageRefetchQuery } from "#/__generated__/core/RiskControlsPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const riskControlsPageQuery = graphql`
  query RiskControlsPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        ...RiskControlsPage_risk
      }
    }
  }
`;

const controlsFragment = graphql`
  fragment RiskControlsPage_risk on Risk
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "ControlOrder", defaultValue: null }
    filter: { type: "ControlFilter", defaultValue: null }
  )
  @refetchable(queryName: "RiskControlsPageRefetchQuery") {
    id
    controls(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "RiskControlsPage_controls") {
      edges {
        node {
          id
          sectionTitle
          name
          framework {
            id
            name
          }
        }
      }
    }
  }
`;

interface RiskControlsPageProps {
  queryRef: PreloadedQuery<RiskControlsPageQuery>;
}

export default function RiskControlsPage(props: RiskControlsPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<RiskControlsPageQuery>(riskControlsPageQuery, props.queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const pagination = usePaginationFragment<
    RiskControlsPageRefetchQuery,
    RiskControlsPage_risk$key
  >(controlsFragment, data.node);
  const controls = pagination.data.controls.edges.map(edge => edge.node);

  return (
    <SortableTable
      {...pagination}
      refetch={
        pagination.refetch as ComponentProps<typeof SortableTable>["refetch"]
      }
    >
      <Thead>
        <Tr>
          <SortableTh field="SECTION_TITLE">{__("Reference")}</SortableTh>
          <Th>{__("Name")}</Th>
        </Tr>
      </Thead>
      <Tbody>
        {controls.length === 0 && (
          <Tr>
            <Td colSpan={2} className="text-center text-txt-secondary">
              {__("No controls linked")}
            </Td>
          </Tr>
        )}
        {controls.map(control => (
          <Tr
            key={control.id}
            to={`/organizations/${organizationId}/frameworks/${control.framework.id}/controls/${control.id}`}
          >
            <Td>
              <span className="inline-flex gap-2 items-center">
                {control.framework.name}
                {" "}
                <Badge size="md">{control.sectionTitle}</Badge>
              </span>
            </Td>
            <Td>{control.name}</Td>
          </Tr>
        ))}
      </Tbody>
    </SortableTable>
  );
}
