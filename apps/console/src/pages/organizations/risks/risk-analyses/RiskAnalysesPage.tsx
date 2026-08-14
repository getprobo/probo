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
import { dateFormat } from "@probo/i18n";
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
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateRiskAnalysisDialog } from "./_components/CreateRiskAnalysisDialog";
import { formatMatrixSize } from "./_components/matrixSize";

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
      action: "core:risk-analysis:create"
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
          name
          description
          period {
            start
            end
          }
          matrixSize {
            rows
            cols
          }
          createdAt
        }
      }
    }
  }
`;

interface RiskAnalysesPageProps {
  queryRef: PreloadedQuery<RiskAnalysesPageQuery>;
}

export default function RiskAnalysesPage({ queryRef }: RiskAnalysesPageProps) {
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();

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
            <SortableTh field="NAME">{t("riskAnalysesPage.columns.name")}</SortableTh>
            <Th>{t("riskAnalysesPage.columns.description")}</Th>
            <Th>{t("riskAnalysesPage.columns.period")}</Th>
            <Th>{t("riskAnalysesPage.columns.matrixSize")}</Th>
            <SortableTh field="CREATED_AT">{t("riskAnalysesPage.columns.created")}</SortableTh>
          </Tr>
        </Thead>
        <Tbody>
          {riskAnalyses.map(ra => (
            <Tr
              key={ra.id}
              to={`/organizations/${organizationId}/risk-management/risk-analyses/${ra.id}`}
            >
              <Td className="font-medium">{ra.name}</Td>
              <Td className="text-txt-secondary truncate max-w-xs">
                {ra.description || "—"}
              </Td>
              <Td className="text-txt-secondary">
                {ra.period
                  ? `${ra.period.start ? dateFormat(i18n.language, ra.period.start) : "—"} – ${ra.period.end ? dateFormat(i18n.language, ra.period.end) : "—"}`
                  : "—"}
              </Td>
              <Td className="text-txt-secondary">
                {formatMatrixSize(ra.matrixSize.rows, ra.matrixSize.cols)}
              </Td>
              <Td className="text-txt-secondary">
                {dateFormat(i18n.language, ra.createdAt)}
              </Td>
            </Tr>
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
