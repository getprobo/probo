// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatDate } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
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
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery(riskAssessmentsPageQuery, queryRef);
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

  usePageTitle(__("Risk Assessments"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Risk Assessments")}
        description={__(
          "Manage risk assessments with scopes, nodes, processes, threats, and scenarios.",
        )}
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
            <SortableTh field="NAME">{__("Name")}</SortableTh>
            <Th>{__("Description")}</Th>
            <SortableTh field="CREATED_AT">{__("Created")}</SortableTh>
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
                {formatDate(ra.createdAt)}
              </Td>
            </Tr>
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
