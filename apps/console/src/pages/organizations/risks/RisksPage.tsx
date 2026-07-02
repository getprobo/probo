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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconPageTextLine,
  IconPlusLarge,
  IconUpload,
  PageHeader,
  RisksChart,
  Tbody,
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
import { useNavigate } from "react-router";

import type { RisksPageFragment$key } from "#/__generated__/core/RisksPageFragment.graphql";
import type { RisksPageQuery } from "#/__generated__/core/RisksPageQuery.graphql";
import type { RisksPageRefetchQuery } from "#/__generated__/core/RisksPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { FormRiskDialog } from "./_components/FormRiskDialog";
import { PublishRiskListDialog } from "./_components/PublishRiskListDialog";
import { RiskRow } from "./_components/RiskRow";

export const risksPageQuery = graphql`
  query RisksPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...RisksPageFragment
    }
  }
`;

const risksFragment = graphql`
  fragment RisksPageFragment on Organization
  @refetchable(queryName: "RisksPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "RiskOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    canCreateRisk: permission(action: "core:risk:create")
    canPublishRisk: permission(action: "core:risk:publish")
    risksDocument {
      id
      defaultApprovers {
        id
      }
    }
    risks(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "RisksPage_risks", filters: []) {
      __id
      edges {
        node {
          id
          name
          inherentLikelihood
          inherentImpact
          residualLikelihood
          residualImpact
          canUpdate: permission(action: "core:risk:update")
          canDelete: permission(action: "core:risk:delete")
          ...RiskRow_risk
        }
      }
    }
  }
`;

export const RisksConnectionKey = "RisksPage_risks";

interface RisksPageProps {
  queryRef: PreloadedQuery<RisksPageQuery>;
}

export default function RisksPage(props: RisksPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const queryData = usePreloadedQuery<RisksPageQuery>(risksPageQuery, props.queryRef);
  const { data: fragmentData, ...pagination } = usePaginationFragment<
    RisksPageRefetchQuery,
    RisksPageFragment$key
  >(risksFragment, queryData.organization);

  const canCreateRisk = fragmentData.canCreateRisk;
  const canPublishRisk = fragmentData.canPublishRisk;
  const risksDocument = fragmentData.risksDocument;
  const risks = fragmentData.risks?.edges.map(edge => edge.node) ?? [];
  const connectionId = fragmentData.risks.__id;

  const chartRisks = risks.map(({
    id,
    name,
    inherentLikelihood,
    inherentImpact,
    residualLikelihood,
    residualImpact,
  }) => ({
    id,
    name,
    inherentLikelihood,
    inherentImpact,
    residualLikelihood,
    residualImpact,
  }));

  const refetch = ({
    order,
  }: {
    order: { direction: string; field: string };
  }) => {
    pagination.refetch(
      {
        order: {
          direction: order.direction as "ASC" | "DESC",
          field: order.field as
          | "NAME"
          | "CATEGORY"
          | "TREATMENT"
          | "INHERENT_RISK_SCORE"
          | "RESIDUAL_RISK_SCORE"
          | "OWNER_FULL_NAME"
          | "CREATED_AT",
        },
      },
      { fetchPolicy: "network-only" },
    );
  };

  usePageTitle(__("Risks"));

  const hasAnyAction = risks.some(
    ({ canDelete, canUpdate }) => canUpdate || canDelete,
  );

  const defaultApproverIds
    = risksDocument?.defaultApprovers?.map(a => a.id) ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Risks")}
        description={__(
          "Risks are potential threats to your organization. Manage them by identifying, assessing, and implementing mitigation measures.",
        )}
      >
        <div className="flex gap-2">
          {risksDocument && (
            <Button
              variant="secondary"
              icon={IconPageTextLine}
              onClick={() => void navigate(
                `/organizations/${organizationId}/documents/${risksDocument.id}`,
              )}
            >
              {__("Document")}
            </Button>
          )}
          {canPublishRisk && (
            <PublishRiskListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={documentId => void navigate(
                `/organizations/${organizationId}/documents/${documentId}`,
              )}
            >
              <Button variant="secondary" icon={IconUpload}>
                {__("Publish")}
              </Button>
            </PublishRiskListDialog>
          )}
          {canCreateRisk && (
            <FormRiskDialog
              connection={connectionId}
              onSuccess={() => {
                pagination.refetch({});
              }}
              trigger={<Button icon={IconPlusLarge}>{__("New Risk")}</Button>}
            />
          )}
        </div>
      </PageHeader>

      <div className="grid grid-cols-2 gap-4">
        <RisksChart
          organizationId={organizationId}
          type="inherent"
          risks={chartRisks}
        />
        <RisksChart
          organizationId={organizationId}
          type="residual"
          risks={chartRisks}
        />
      </div>
      <SortableTable {...pagination} refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Risk name")}</SortableTh>
            <SortableTh field="CATEGORY">{__("Category")}</SortableTh>
            <SortableTh field="TREATMENT">{__("Treatment")}</SortableTh>
            <SortableTh field="INHERENT_RISK_SCORE">
              {__("Initial Risk")}
            </SortableTh>
            <SortableTh field="RESIDUAL_RISK_SCORE">
              {__("Residual Risk")}
            </SortableTh>
            <SortableTh field="OWNER_FULL_NAME">{__("Owner")}</SortableTh>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {risks.map(risk => (
            <RiskRow
              key={risk.id}
              riskKey={risk}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
