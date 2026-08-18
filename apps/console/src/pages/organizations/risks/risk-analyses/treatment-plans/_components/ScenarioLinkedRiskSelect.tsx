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

import { Option } from "@probo/ui";
import { type Control, type FieldValues, type Path } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { ScenarioLinkedRiskSelect_analysis$key } from "#/__generated__/core/ScenarioLinkedRiskSelect_analysis.graphql";
import type { ScenarioLinkedRiskSelectQuery } from "#/__generated__/core/ScenarioLinkedRiskSelectQuery.graphql";
import type { ScenarioLinkedRiskSelectRefetchQuery } from "#/__generated__/core/ScenarioLinkedRiskSelectRefetchQuery.graphql";
import { ControlledField } from "#/components/form/ControlledField";

import { TREATMENT_PLAN_PAGE_SIZE, useLoadAllPages } from "../_lib/useLoadAllPages";

export const scenarioLinkedRiskSelectQuery = graphql`
  query ScenarioLinkedRiskSelectQuery($riskAnalysisId: ID!) {
    node(id: $riskAnalysisId) {
      __typename
      ... on RiskAnalysis {
        ...ScenarioLinkedRiskSelect_analysis
      }
    }
  }
`;

const analysisFragment = graphql`
  fragment ScenarioLinkedRiskSelect_analysis on RiskAnalysis
  @refetchable(queryName: "ScenarioLinkedRiskSelectRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    scenarioRisks(
      first: $first
      after: $after
      last: $last
      before: $before
    ) @connection(key: "ScenarioLinkedRiskSelect_scenarioRisks") {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`;

interface ScenarioLinkedRiskSelectProps<TFieldValues extends FieldValues & { riskId: string }> {
  control: Control<TFieldValues>;
  error?: string;
  queryRef: PreloadedQuery<ScenarioLinkedRiskSelectQuery>;
}

export function ScenarioLinkedRiskSelect<
  TFieldValues extends FieldValues & { riskId: string },
>({
  control,
  error,
  queryRef,
}: ScenarioLinkedRiskSelectProps<TFieldValues>) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<ScenarioLinkedRiskSelectQuery>(
    scenarioLinkedRiskSelectQuery,
    queryRef,
  );
  const analysisRef: ScenarioLinkedRiskSelect_analysis$key | null
    = data.node?.__typename === "RiskAnalysis" ? data.node : null;
  const { data: analysis, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    ScenarioLinkedRiskSelectRefetchQuery,
    ScenarioLinkedRiskSelect_analysis$key
  >(analysisFragment, analysisRef);
  useLoadAllPages(hasNext, isLoadingNext, loadNext, TREATMENT_PLAN_PAGE_SIZE);

  const risks = [...(analysis?.scenarioRisks?.edges ?? [])]
    .map(edge => edge.node)
    .sort((a, b) => a.name.localeCompare(b.name));

  if (risks.length === 0 && !hasNext && !isLoadingNext) {
    return (
      <p className="text-sm text-txt-secondary">
        {t("createTreatmentPlanDialog.emptyEligible")}
      </p>
    );
  }

  return (
    <ControlledField
      control={control}
      error={error}
      help={t("createTreatmentPlanDialog.hint")}
      label={t("createTreatmentPlanDialog.fields.risk")}
      name={"riskId" as Path<TFieldValues>}
      placeholder={t("createTreatmentPlanDialog.placeholders.risk")}
      type="select"
    >
      {risks.map(risk => (
        <Option key={risk.id} value={risk.id}>{risk.name}</Option>
      ))}
    </ControlledField>
  );
}
