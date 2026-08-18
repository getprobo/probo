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

import { Button, IconChevronDown, Spinner, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { RiskTreatmentPlansPageFragment$key } from "#/__generated__/core/RiskTreatmentPlansPageFragment.graphql";
import type { RiskTreatmentPlansPageQuery } from "#/__generated__/core/RiskTreatmentPlansPageQuery.graphql";
import type { RiskTreatmentPlansPageRefetchQuery } from "#/__generated__/core/RiskTreatmentPlansPageRefetchQuery.graphql";

import { RiskTreatmentPlanListItem } from "./_components/RiskTreatmentPlanListItem";

const PAGE_SIZE = 50;

export const riskTreatmentPlansPageQuery = graphql`
  query RiskTreatmentPlansPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        ...RiskTreatmentPlansPageFragment
      }
    }
  }
`;

const riskTreatmentPlansPageFragment = graphql`
  fragment RiskTreatmentPlansPageFragment on Risk
  @refetchable(queryName: "RiskTreatmentPlansPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    id
    treatmentPlans(first: $first, after: $after, last: $last, before: $before)
      @connection(key: "RiskTreatmentPlansPage_treatmentPlans", filters: []) {
      __id
      edges {
        node {
          id
          ...RiskTreatmentPlanListItem_treatmentPlan
        }
      }
    }
  }
`;

interface RiskTreatmentPlansPageProps {
  queryRef: PreloadedQuery<RiskTreatmentPlansPageQuery>;
}

export default function RiskTreatmentPlansPage({ queryRef }: RiskTreatmentPlansPageProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<RiskTreatmentPlansPageQuery>(
    riskTreatmentPlansPageQuery,
    queryRef,
  );
  const riskRef: RiskTreatmentPlansPageFragment$key | null
    = data.node?.__typename === "Risk" ? data.node : null;
  const { data: risk, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    RiskTreatmentPlansPageRefetchQuery,
    RiskTreatmentPlansPageFragment$key
  >(riskTreatmentPlansPageFragment, riskRef);
  if (!risk) {
    throw new Error("Risk not found");
  }
  const plans = risk.treatmentPlans.edges.map(edge => edge.node);
  const connectionId = risk.treatmentPlans.__id;

  return (
    <div className="space-y-4">
      <Table>
        <Thead>
          <Tr>
            <Th>{t("riskTreatmentPlansPage.columns.analysis")}</Th>
            <Th>{t("riskTreatmentPlansPage.columns.treatment")}</Th>
            <Th>{t("riskTreatmentPlansPage.columns.owner")}</Th>
            <Th>{t("riskTreatmentPlansPage.columns.scores")}</Th>
            <Th className="w-12" />
          </Tr>
        </Thead>
        <Tbody>
          {plans.length === 0 && (
            <Tr>
              <Td colSpan={5} className="text-center text-txt-secondary">
                {t("riskTreatmentPlansPage.empty")}
              </Td>
            </Tr>
          )}
          {plans.map(plan => (
            <RiskTreatmentPlanListItem
              key={plan.id}
              treatmentPlanKey={plan}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </Table>
      {hasNext && (
        <Button
          variant="tertiary"
          onClick={() => loadNext(PAGE_SIZE)}
          className="mx-auto"
          disabled={isLoadingNext}
          icon={isLoadingNext ? Spinner : IconChevronDown}
        >
          {t("sortableTable.actions.showMore")}
        </Button>
      )}
    </div>
  );
}
