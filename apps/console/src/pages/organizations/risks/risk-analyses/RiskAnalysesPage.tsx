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

import { usePageTitle } from "@probo/hooks";
import {
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { RiskAnalysesPageFragment$key } from "#/__generated__/core/RiskAnalysesPageFragment.graphql";
import type { RiskAnalysesPageQuery } from "#/__generated__/core/RiskAnalysesPageQuery.graphql";
import type { RiskAnalysesPageRefetchQuery } from "#/__generated__/core/RiskAnalysesPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { CreateRiskAnalysisDialog } from "./_components/CreateRiskAnalysisDialog";
import { RiskAnalysisListItem } from "./_components/RiskAnalysisListItem";

export const riskAnalysesPageQuery = graphql`
  query RiskAnalysesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...RiskAnalysesPageFragment
    }
  }
`;

const riskAnalysesFragment = graphql`
  fragment RiskAnalysesPageFragment on Organization
  @refetchable(queryName: "RiskAnalysesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "RiskAnalysisOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    canCreateRiskAnalysis: permission(
      action: "risk-management:risk-analysis:create"
    )
    riskAnalyses(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    )
      @connection(
        key: "RiskAnalysesPage_riskAnalyses"
        filters: []
      ) {
      __id
      edges {
        node {
          id
          ...RiskAnalysisListItem_riskAnalysis
        }
      }
    }
  }
`;

interface RiskAnalysesPageProps {
  queryRef: PreloadedQuery<RiskAnalysesPageQuery>;
}

export default function RiskAnalysesPage({ queryRef }: RiskAnalysesPageProps) {
  const { t } = useTranslation();

  const data = usePreloadedQuery<RiskAnalysesPageQuery>(riskAnalysesPageQuery, queryRef);
  const { data: fragmentData, ...pagination } = usePaginationFragment<
    RiskAnalysesPageRefetchQuery,
    RiskAnalysesPageFragment$key
  >(riskAnalysesFragment, data.organization);

  const riskAnalyses
    = fragmentData.riskAnalyses?.edges.map(edge => edge.node) ?? [];
  const connectionId = fragmentData.riskAnalyses.__id;
  const canCreate = fragmentData.canCreateRiskAnalysis;

  const refetch = ({
    order,
  }: {
    order: { direction: string; field: string };
  }) => {
    pagination.refetch(
      {
        order: {
          direction: order.direction as "ASC" | "DESC",
          field: order.field as "NAME" | "CREATED_AT",
        },
      },
      { fetchPolicy: "network-only" },
    );
  };

  usePageTitle(t("riskAnalysesPage.title"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("riskAnalysesPage.title")}
        description={t("riskAnalysesPage.description")}
      >
        {canCreate && (
          <CreateRiskAnalysisDialog
            connectionId={connectionId}
          />
        )}
      </PageHeader>

      <SortableTable {...pagination} refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="NAME" className="w-72 min-w-48">
              {t("riskAnalysesPage.columns.name")}
            </SortableTh>
            <Th className="w-full">{t("riskAnalysesPage.columns.description")}</Th>
            <Th>{t("riskAnalysesPage.columns.period")}</Th>
            <Th>{t("riskAnalysesPage.columns.matrixSize")}</Th>
            <SortableTh field="CREATED_AT">
              {t("riskAnalysesPage.columns.created")}
            </SortableTh>
            <Th />
          </Tr>
        </Thead>
        <Tbody>
          {riskAnalyses.length === 0 && (
            <Tr>
              <Td colSpan={6} className="text-center text-txt-secondary">
                {t("riskAnalysesPage.empty")}
              </Td>
            </Tr>
          )}
          {riskAnalyses.map(ra => (
            <RiskAnalysisListItem
              key={ra.id}
              riskAnalysisKey={ra}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
