// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { Badge, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
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
          <SortableTh field="SECTION_TITLE">{t("riskControlsPage.columns.reference")}</SortableTh>
          <Th>{t("riskControlsPage.columns.name")}</Th>
        </Tr>
      </Thead>
      <Tbody>
        {controls.length === 0 && (
          <Tr>
            <Td colSpan={2} className="text-center text-txt-secondary">
              {t("riskControlsPage.empty")}
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
