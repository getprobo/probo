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

import { Button, IconCrossLargeX, IconWarning, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { RiskAnalysisPlansSection_analysis$key } from "#/__generated__/core/RiskAnalysisPlansSection_analysis.graphql";
import type { RiskAnalysisPlansSectionRefetchQuery } from "#/__generated__/core/RiskAnalysisPlansSectionRefetchQuery.graphql";
import { type Order, SortableContext, SortableTh } from "#/components/SortableTable";

import { matchesMatrixCell } from "../_lib/matrixCell";
import { useMatrixCellFilter } from "../_lib/useMatrixCellFilter";
import { LinkedRiskListItem } from "../treatment-plans/_components/LinkedRiskListItem";
import { TreatmentPlanListItem } from "../treatment-plans/_components/TreatmentPlanListItem";

import { AnalysisSectionError } from "./AnalysisSectionError";
import { RiskAnalysisMatrices } from "./RiskAnalysisMatrices";

export const riskAnalysisPlansSectionFragment = graphql`
  fragment RiskAnalysisPlansSection_analysis on RiskAnalysis
  @refetchable(queryName: "RiskAnalysisPlansSectionRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    id
    matrixSize {
      rows
      cols
    }
    ...RiskAnalysisMatrices_analysis @arguments(first: $first, after: $after)
    ...LinkedRiskListItem_analysis
    treatmentPlans(first: $first, after: $after)
      @connection(key: "RiskAnalysisPlansSection_treatmentPlans", filters: []) {
      __id
      edges {
        node {
          id
          inherentLikelihood
          inherentImpact
          netLikelihood
          netImpact
          residualLikelihood
          residualImpact
          risk {
            id
            category
          }
          ...TreatmentPlanListItem_treatmentPlan
        }
      }
    }
    scenarioRisks(first: 500) {
      edges {
        node {
          id
          category
          ...LinkedRiskListItem_risk
        }
      }
    }
  }
`;

interface RiskAnalysisPlansSectionProps {
  analysisKey: RiskAnalysisPlansSection_analysis$key;
}

export function RiskAnalysisPlansSection({ analysisKey }: RiskAnalysisPlansSectionProps) {
  return (
    <AnalysisSectionError>
      <RiskAnalysisPlansSectionContent analysisKey={analysisKey} />
    </AnalysisSectionError>
  );
}

function RiskAnalysisPlansSectionContent({ analysisKey }: RiskAnalysisPlansSectionProps) {
  const { t } = useTranslation();
  const { cell, clear } = useMatrixCellFilter();
  const [order, setOrder] = useState<Order>({ direction: "DESC", field: "" });
  const [ra, refetch] = useRefetchableFragment<
    RiskAnalysisPlansSectionRefetchQuery,
    RiskAnalysisPlansSection_analysis$key
  >(riskAnalysisPlansSectionFragment, analysisKey);

  if (!ra.id || !ra.matrixSize) {
    return null;
  }

  const matrixSize = { rows: ra.matrixSize.rows, cols: ra.matrixSize.cols };
  const plans = ra.treatmentPlans?.edges.map(edge => edge.node) ?? [];
  const connectionId = ra.treatmentPlans?.__id ?? "";
  const plannedRiskIds = new Set(plans.map(plan => plan.risk.id));
  const unplannedRisks = (ra.scenarioRisks?.edges.map(edge => edge.node) ?? [])
    .filter(risk => !plannedRiskIds.has(risk.id));
  const visiblePlans = cell
    ? plans.filter(plan => matchesMatrixCell({
        inherentLikelihood: plan.inherentLikelihood,
        inherentImpact: plan.inherentImpact,
        netLikelihood: plan.netLikelihood,
        netImpact: plan.netImpact,
        residualLikelihood: plan.residualLikelihood,
        residualImpact: plan.residualImpact,
      }, cell))
    : plans;
  const visibleUnplannedRisks = cell ? [] : unplannedRisks;

  const reload = () => {
    refetch({}, { fetchPolicy: "network-only" });
  };

  const hasLinkedRisks = plans.length > 0 || unplannedRisks.length > 0;
  const isEmpty = visiblePlans.length === 0 && visibleUnplannedRisks.length === 0;
  const rows = sortPlanRows(visiblePlans, visibleUnplannedRisks, order);

  return (
    <div className="space-y-4">
      {unplannedRisks.length > 0 && (
        <div className="flex items-center gap-3 rounded-lg bg-warning px-4 py-3 text-sm text-txt-warning">
          <IconWarning size={16} className="shrink-0" />
          {t("riskAnalysisTreatmentPlansPage.unplannedWarning", {
            count: unplannedRisks.length,
          })}
        </div>
      )}
      <RiskAnalysisMatrices analysisKey={ra} />
      {cell && (
        <div className="flex justify-end">
          <Button variant="secondary" icon={IconCrossLargeX} onClick={clear}>
            {t("riskAnalysisTreatmentPlansPage.filter", {
              state: t(`riskAnalysisTreatmentPlansPage.states.${cell.type}`),
              likelihood: cell.likelihood,
              impact: cell.impact,
            })}
          </Button>
        </div>
      )}
      <SortableContext value={{ order, changeOrder: setOrder }}>
        <Table>
          <Thead>
            <Tr>
              <Th>{t("riskAnalysisTreatmentPlansPage.columns.risk")}</Th>
              <SortableTh field="CATEGORY" className="w-px pr-6">
                {t("riskAnalysisTreatmentPlansPage.columns.category")}
              </SortableTh>
              <Th className="w-px pr-6">{t("riskAnalysisTreatmentPlansPage.columns.treatment")}</Th>
              <Th className="w-px pr-6">{t("riskAnalysisTreatmentPlansPage.columns.owner")}</Th>
              <Th className="w-px pr-6">{t("riskAnalysisTreatmentPlansPage.columns.scores")}</Th>
              <Th className="w-px pr-6">{t("riskAnalysisTreatmentPlansPage.columns.progress")}</Th>
              <Th width={48} className="w-px" />
            </Tr>
          </Thead>
          <Tbody>
            {isEmpty && (
              <Tr>
                <Td
                  colSpan={7}
                  className="text-center text-txt-secondary"
                >
                  {hasLinkedRisks
                    ? t("riskAnalysisTreatmentPlansPage.emptyFilter")
                    : t("riskAnalysisTreatmentPlansPage.empty")}
                </Td>
              </Tr>
            )}
            {rows.map(row =>
              row.kind === "plan"
                ? (
                    <TreatmentPlanListItem
                      key={row.plan.id}
                      treatmentPlanKey={row.plan}
                      connectionId={connectionId}
                      matrixSize={matrixSize}
                      onChanged={reload}
                    />
                  )
                : (
                    <LinkedRiskListItem
                      key={row.risk.id}
                      riskKey={row.risk}
                      analysisKey={ra}
                      connectionId={connectionId}
                      matrixSize={matrixSize}
                      onCreated={reload}
                    />
                  ),
            )}
          </Tbody>
        </Table>
      </SortableContext>
    </div>
  );
}

function sortPlanRows<
  TPlan extends { id: string; risk: { category: string } },
  TRisk extends { id: string; category: string },
>(plans: TPlan[], risks: TRisk[], order: Order) {
  const combined = [
    ...plans.map(plan => ({ kind: "plan" as const, plan, category: plan.risk.category })),
    ...risks.map(risk => ({ kind: "unplanned" as const, risk, category: risk.category })),
  ];
  if (order.field !== "CATEGORY") {
    return combined;
  }
  const direction = order.direction === "ASC" ? 1 : -1;
  return [...combined].sort((a, b) =>
    direction * a.category.localeCompare(b.category, undefined, { sensitivity: "base" }),
  );
}
