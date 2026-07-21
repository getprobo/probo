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

import type { RiskAssessmentsPageFragment$key } from "#/__generated__/core/RiskAssessmentsPageFragment.graphql";
import type { RiskAssessmentsPageQuery } from "#/__generated__/core/RiskAssessmentsPageQuery.graphql";
import type { RiskAssessmentsPageRefetchQuery } from "#/__generated__/core/RiskAssessmentsPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateRiskAssessmentDialog } from "./_components/CreateRiskAssessmentDialog";

export const riskAssessmentsPageQuery = graphql`
  query RiskAssessmentsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...RiskAssessmentsPageFragment
    }
  }
`;

const riskAssessmentsFragment = graphql`
  fragment RiskAssessmentsPageFragment on Organization
  @refetchable(queryName: "RiskAssessmentsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "RiskAssessmentOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    canCreateRiskAssessment: permission(
      action: "core:risk-assessment:create"
    )
    riskAssessments(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    )
      @connection(
        key: "RiskAssessmentsPage_riskAssessments"
        filters: []
      ) {
      __id
      edges {
        node {
          id
          name
          description
          createdAt
        }
      }
    }
  }
`;

interface RiskAssessmentsPageProps {
  queryRef: PreloadedQuery<RiskAssessmentsPageQuery>;
}

export default function RiskAssessmentsPage({ queryRef }: RiskAssessmentsPageProps) {
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery<RiskAssessmentsPageQuery>(riskAssessmentsPageQuery, queryRef);
  const { data: fragmentData, ...pagination } = usePaginationFragment<
    RiskAssessmentsPageRefetchQuery,
    RiskAssessmentsPageFragment$key
  >(riskAssessmentsFragment, data.organization);

  const riskAssessments
    = fragmentData.riskAssessments?.edges.map(edge => edge.node) ?? [];
  const connectionId = fragmentData.riskAssessments.__id;
  const canCreate = fragmentData.canCreateRiskAssessment;

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

  usePageTitle(t("riskAssessmentsPage.title"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("riskAssessmentsPage.title")}
        description={t("riskAssessmentsPage.description")}
      >
        {canCreate && (
          <CreateRiskAssessmentDialog
            connectionId={connectionId}
          />
        )}
      </PageHeader>

      <SortableTable {...pagination} refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{t("riskAssessmentsPage.columns.name")}</SortableTh>
            <Th>{t("riskAssessmentsPage.columns.description")}</Th>
            <SortableTh field="CREATED_AT">{t("riskAssessmentsPage.columns.created")}</SortableTh>
          </Tr>
        </Thead>
        <Tbody>
          {riskAssessments.map(ra => (
            <Tr
              key={ra.id}
              to={`/organizations/${organizationId}/risk-assessments/${ra.id}`}
            >
              <Td className="font-medium">{ra.name}</Td>
              <Td className="text-txt-secondary truncate max-w-xs">
                {ra.description || "—"}
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
