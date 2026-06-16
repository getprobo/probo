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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
import { graphql, type PreloadedQuery, useRefetchableFragment, usePreloadedQuery } from "react-relay";

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
  const queryData = usePreloadedQuery(
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
  const { __ } = useTranslate();
  const [expanded, setExpanded] = useState<string | null>(null);

  usePageTitle(data.name + " - " + __("Risk Assessments"));

  if (assessments.length === 0) {
    return (
      <div className="text-center text-sm py-6 text-txt-secondary flex flex-col items-center gap-2">
        {__("No risk assessments found")}
        {data.canCreateRiskAssessment && (
          <CreateRiskAssessmentDialog
            thirdPartyId={data.id}
            connection={data.riskAssessments.__id}
          >
            <Button icon={IconPlusLarge} variant="secondary">
              {__("Add Risk Assessment")}
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
              <SortableTh field="CREATED_AT">{__("Created At")}</SortableTh>
              <SortableTh field="EXPIRES_AT">{__("Expires")}</SortableTh>
              <Th>{__("Data sensitivity")}</Th>
              <Th>{__("Business impact")}</Th>
            </Tr>
          </Thead>
          <Tbody>
            {data.canCreateRiskAssessment && (
              <CreateRiskAssessmentDialog
                thirdPartyId={data.id}
                connection={data.riskAssessments.__id}
              >
                <TrButton colspan={5} onClick={() => {}}>
                  {__("Add Risk Assessment")}
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
