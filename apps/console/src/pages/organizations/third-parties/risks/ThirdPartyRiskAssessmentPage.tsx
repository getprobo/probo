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
import {
  Button,
  IconPlusLarge,
  Tbody,
  Th,
  Thead,
  Tr,
  TrButton,
} from "@probo/ui";
import type { ComponentProps } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery, useRefetchableFragment } from "react-relay";

import type { ThirdPartyRiskAssessmentPageFragment$key } from "#/__generated__/core/ThirdPartyRiskAssessmentPageFragment.graphql";
import type { ThirdPartyRiskAssessmentPageQuery } from "#/__generated__/core/ThirdPartyRiskAssessmentPageQuery.graphql";
import type { ThirdPartyRiskAssessmentPageRefetchQuery } from "#/__generated__/core/ThirdPartyRiskAssessmentPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { CreateRiskAssessmentDialog } from "../_components/CreateRiskAssessmentDialog";

import { ThirdPartyRiskAssessmentRow } from "./_components/ThirdPartyRiskAssessmentRow";

const riskAssessmentsFragment = graphql`
  fragment ThirdPartyRiskAssessmentPageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyRiskAssessmentPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyRiskAssessmentOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    name
    canCreateRiskAssessment: permission(
      action: "core:thirdParty-risk-assessment:create"
    )
    riskAssessments(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "ThirdPartyRiskAssessmentPage_riskAssessments") {
      __id
      edges {
        node {
          id
          ...ThirdPartyRiskAssessmentRow_assessment
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export const thirdPartyRiskAssessmentPageQuery = graphql`
  query ThirdPartyRiskAssessmentPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        ...ThirdPartyRiskAssessmentPageFragment
      }
    }
  }
`;

interface ThirdPartyRiskAssessmentPageProps {
  queryRef: PreloadedQuery<ThirdPartyRiskAssessmentPageQuery>;
}

export default function ThirdPartyRiskAssessmentPage(
  props: ThirdPartyRiskAssessmentPageProps,
) {
  const queryData = usePreloadedQuery<ThirdPartyRiskAssessmentPageQuery>(
    thirdPartyRiskAssessmentPageQuery,
    props.queryRef,
  );
  if (queryData.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }

  const [data, refetch] = useRefetchableFragment<
    ThirdPartyRiskAssessmentPageRefetchQuery,
    ThirdPartyRiskAssessmentPageFragment$key
  >(riskAssessmentsFragment, queryData.node);

  const assessments = data.riskAssessments.edges.map(edge => edge.node);
  const { t } = useTranslation();
  const [expanded, setExpanded] = useState<string | null>(null);

  usePageTitle(t("thirdPartyRiskAssessmentPage.pageTitle", { name: data.name }));

  if (assessments.length === 0) {
    return (
      <div className="text-center text-sm py-6 text-txt-secondary flex flex-col items-center gap-2">
        {t("thirdPartyRiskAssessmentPage.empty")}
        {data.canCreateRiskAssessment && (
          <CreateRiskAssessmentDialog
            thirdPartyId={data.id}
            connection={data.riskAssessments.__id}
          >
            <Button icon={IconPlusLarge} variant="secondary">
              {t("thirdPartyRiskAssessmentPage.actions.add")}
            </Button>
          </CreateRiskAssessmentDialog>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6 relative">
      <div className="overflow-x-auto">
        <SortableTable
          refetch={refetch as ComponentProps<typeof SortableTable>["refetch"]}
        >
          <Thead>
            <Tr>
              <SortableTh field="CREATED_AT">{t("thirdPartyRiskAssessmentPage.columns.createdAt")}</SortableTh>
              <SortableTh field="EXPIRES_AT">{t("thirdPartyRiskAssessmentPage.columns.expires")}</SortableTh>
              <Th>{t("thirdPartyRiskAssessmentPage.columns.dataSensitivity")}</Th>
              <Th>{t("thirdPartyRiskAssessmentPage.columns.businessImpact")}</Th>
            </Tr>
          </Thead>
          <Tbody>
            {data.canCreateRiskAssessment && (
              <CreateRiskAssessmentDialog
                thirdPartyId={data.id}
                connection={data.riskAssessments.__id}
              >
                <TrButton colspan={4} onClick={() => {}}>
                  {t("thirdPartyRiskAssessmentPage.actions.add")}
                </TrButton>
              </CreateRiskAssessmentDialog>
            )}
            {assessments.map(assessment => (
              <ThirdPartyRiskAssessmentRow
                key={assessment.id}
                assessmentKey={assessment}
                isExpanded={expanded === assessment.id}
                onClick={() =>
                  setExpanded(prev =>
                    prev === assessment.id ? null : assessment.id,
                  )}
              />
            ))}
          </Tbody>
        </SortableTable>
      </div>
    </div>
  );
}
