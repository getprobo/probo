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

import {
  Button,
  IconChevronDown,
  IconCrossLargeX,
  IconWarning,
  Spinner,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { Suspense, useEffect, useMemo, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, usePaginationFragment } from "react-relay";

import type { RiskAnalysisPlansSection_analysis$key } from "#/__generated__/core/RiskAnalysisPlansSection_analysis.graphql";
import type { RiskAnalysisPlansSection_plans$key } from "#/__generated__/core/RiskAnalysisPlansSection_plans.graphql";
import type { RiskAnalysisPlansSection_unplanned$key } from "#/__generated__/core/RiskAnalysisPlansSection_unplanned.graphql";
import type { RiskAnalysisPlansSectionPlansQuery } from "#/__generated__/core/RiskAnalysisPlansSectionPlansQuery.graphql";
import type { RiskAnalysisPlansSectionUnplannedQuery } from "#/__generated__/core/RiskAnalysisPlansSectionUnplannedQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { type Order, SortableContext, SortableTh } from "#/components/SortableTable";

import { treatmentPlanFilterFromCell } from "../_lib/matrixCell";
import { useMatrixCellFilter } from "../_lib/useMatrixCellFilter";
import { LinkedRiskListItem } from "../treatment-plans/_components/LinkedRiskListItem";
import { TreatmentPlanListItem } from "../treatment-plans/_components/TreatmentPlanListItem";
import { TREATMENT_PLAN_PAGE_SIZE } from "../treatment-plans/_lib/treatmentPlanPageSize";

import { AnalysisSectionError } from "./AnalysisSectionError";
import { RiskAnalysisMatrices } from "./RiskAnalysisMatrices";

const riskAnalysisPlansSectionPlansFragment = graphql`
  fragment RiskAnalysisPlansSection_plans on RiskAnalysis
  @refetchable(queryName: "RiskAnalysisPlansSectionPlansQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "TreatmentPlanFilter", defaultValue: null }
    orderBy: { type: "TreatmentPlanOrder", defaultValue: null }
  ) {
    treatmentPlans(
      first: $first
      after: $after
      last: $last
      before: $before
      filter: $filter
      orderBy: $orderBy
    )
      @connection(key: "RiskAnalysisPlansSection_treatmentPlans", filters: ["filter", "orderBy"]) {
      __id
      totalCount
      edges {
        node {
          id
          ...TreatmentPlanListItem_treatmentPlan
        }
      }
    }
  }
`;

const riskAnalysisPlansSectionUnplannedFragment = graphql`
  fragment RiskAnalysisPlansSection_unplanned on RiskAnalysis
  @refetchable(queryName: "RiskAnalysisPlansSectionUnplannedQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    orderBy: { type: "RiskOrder", defaultValue: null }
  ) {
    scenarioRisks(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
    ) @connection(key: "RiskAnalysisPlansSection_scenarioRisks", filters: ["orderBy"]) {
      totalCount
      edges {
        node {
          id
          ...LinkedRiskListItem_risk
        }
      }
    }
  }
`;

export const riskAnalysisPlansSectionFragment = graphql`
  fragment RiskAnalysisPlansSection_analysis on RiskAnalysis
  @argumentDefinitions(
    filter: { type: "TreatmentPlanFilter", defaultValue: null }
  ) {
    id
    matrixSize {
      rows
      cols
    }
    ...RiskAnalysisMatrices_analysis
    ...LinkedRiskListItem_analysis
    ...RiskAnalysisPlansSection_plans @arguments(filter: $filter)
    ...RiskAnalysisPlansSection_unplanned
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
  const analysis = useFragment(riskAnalysisPlansSectionFragment, analysisKey);

  if (!analysis.id || !analysis.matrixSize) {
    return null;
  }

  return (
    <div className="space-y-4">
      <RiskAnalysisMatrices analysisKey={analysis} />
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
      <Suspense fallback={<LinkCardSkeleton />}>
        <RiskAnalysisPlansTable analysisKey={analysisKey} />
      </Suspense>
    </div>
  );
}

function RiskAnalysisPlansTable({
  analysisKey,
}: {
  analysisKey: RiskAnalysisPlansSection_analysis$key;
}) {
  const { t } = useTranslation();
  const { cell } = useMatrixCellFilter();
  const [order, setOrder] = useState<Order>({ direction: "DESC", field: "" });
  const [, startTransition] = useTransition();
  const analysis = useFragment(riskAnalysisPlansSectionFragment, analysisKey);
  const {
    data: plansData,
    hasNext: plansHasNext,
    isLoadingNext: plansLoadingNext,
    loadNext: loadNextPlans,
    refetch: refetchPlans,
  } = usePaginationFragment<
    RiskAnalysisPlansSectionPlansQuery,
    RiskAnalysisPlansSection_plans$key
  >(riskAnalysisPlansSectionPlansFragment, analysis);
  const {
    data: unplannedData,
    hasNext: unplannedHasNext,
    isLoadingNext: unplannedLoadingNext,
    loadNext: loadNextUnplanned,
    refetch: refetchUnplanned,
  } = usePaginationFragment<
    RiskAnalysisPlansSectionUnplannedQuery,
    RiskAnalysisPlansSection_unplanned$key
  >(riskAnalysisPlansSectionUnplannedFragment, analysis);
  const filter = useMemo(
    () => treatmentPlanFilterFromCell(cell),
    [cell],
  );
  const orderBy = useMemo(
    () => categoryOrder(order),
    [order],
  );
  const skipFirstPlansRefetch = useRef(true);
  const skipFirstUnplannedRefetch = useRef(true);

  useEffect(() => {
    if (skipFirstPlansRefetch.current) {
      skipFirstPlansRefetch.current = false;
      return;
    }
    startTransition(() => {
      refetchPlans({ filter, orderBy });
    });
  }, [filter, orderBy, refetchPlans, startTransition]);

  useEffect(() => {
    if (skipFirstUnplannedRefetch.current) {
      skipFirstUnplannedRefetch.current = false;
      return;
    }
    startTransition(() => {
      refetchUnplanned({ orderBy });
    });
  }, [orderBy, refetchUnplanned, startTransition]);

  if (!analysis.matrixSize) {
    return null;
  }

  const matrixSize = { rows: analysis.matrixSize.rows, cols: analysis.matrixSize.cols };
  const plans = plansData.treatmentPlans?.edges.map(edge => edge.node) ?? [];
  const connectionId = plansData.treatmentPlans?.__id ?? "";
  const unplannedRisks = unplannedData.scenarioRisks?.edges.map(edge => edge.node) ?? [];
  const unplannedTotal = unplannedData.scenarioRisks?.totalCount ?? 0;
  const visibleUnplannedRisks = cell ? [] : unplannedRisks;
  const showMore = cell ? plansHasNext : plansHasNext || unplannedHasNext;
  const loadingMore = plansLoadingNext || unplannedLoadingNext;

  const reload = () => {
    startTransition(() => {
      refetchPlans({ filter, orderBy }, { fetchPolicy: "network-only" });
      refetchUnplanned({ orderBy }, { fetchPolicy: "network-only" });
    });
  };

  const isEmpty = plans.length === 0 && visibleUnplannedRisks.length === 0;
  const rows = [
    ...plans.map(plan => ({ kind: "plan" as const, plan })),
    ...visibleUnplannedRisks.map(risk => ({ kind: "unplanned" as const, risk })),
  ];

  return (
    <div className="space-y-4">
      {unplannedTotal > 0 && (
        <div className="flex items-center gap-3 rounded-lg bg-warning px-4 py-3 text-sm text-txt-warning">
          <IconWarning size={16} className="shrink-0" />
          {t("riskAnalysisTreatmentPlansPage.unplannedWarning", {
            count: unplannedTotal,
          })}
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
                  {cell
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
                      analysisKey={analysis}
                      connectionId={connectionId}
                      matrixSize={matrixSize}
                      onCreated={reload}
                    />
                  ),
            )}
          </Tbody>
        </Table>
      </SortableContext>
      {showMore && (
        <Button
          variant="tertiary"
          onClick={() => {
            if (plansHasNext) {
              loadNextPlans(TREATMENT_PLAN_PAGE_SIZE);
              return;
            }
            loadNextUnplanned(TREATMENT_PLAN_PAGE_SIZE);
          }}
          className="mx-auto"
          disabled={loadingMore}
          icon={loadingMore ? Spinner : IconChevronDown}
        >
          {t("sortableTable.actions.showMore")}
        </Button>
      )}
    </div>
  );
}

function categoryOrder(order: Order) {
  if (order.field !== "CATEGORY") {
    return null;
  }

  return {
    field: "CATEGORY" as const,
    direction: order.direction,
  };
}
