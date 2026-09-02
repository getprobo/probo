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
import { type ReactNode, Suspense, useEffect, useMemo, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment, useRefetchableFragment } from "react-relay";

import type {
  RiskAnalysisPlansSection_analysis$data,
  RiskAnalysisPlansSection_analysis$key,
} from "#/__generated__/core/RiskAnalysisPlansSection_analysis.graphql";
import type { RiskAnalysisPlansSection_plans$key } from "#/__generated__/core/RiskAnalysisPlansSection_plans.graphql";
import type {
  RiskAnalysisPlansSection_unplanned$data,
  RiskAnalysisPlansSection_unplanned$key,
} from "#/__generated__/core/RiskAnalysisPlansSection_unplanned.graphql";
import type { RiskAnalysisPlansSectionPlansQuery } from "#/__generated__/core/RiskAnalysisPlansSectionPlansQuery.graphql";
import type { RiskAnalysisPlansSectionRefetchQuery } from "#/__generated__/core/RiskAnalysisPlansSectionRefetchQuery.graphql";
import type { RiskAnalysisPlansSectionUnplannedQuery } from "#/__generated__/core/RiskAnalysisPlansSectionUnplannedQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { type Order, SortableContext, SortableTh } from "#/components/SortableTable";

import { asOfDateBounds, clampDateInput, matrixAsOf } from "../_lib/matrixAsOf";
import { treatmentPlanFilterFromCell } from "../_lib/matrixCell";
import { useMatrixAsOf } from "../_lib/useMatrixAsOf";
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
    asOf: { type: "Datetime", defaultValue: null }
  ) {
    treatmentPlans(
      first: $first
      after: $after
      last: $last
      before: $before
      filter: $filter
      orderBy: $orderBy
      asOf: $asOf
    )
      @connection(key: "RiskAnalysisPlansSection_treatmentPlans", filters: ["filter", "orderBy", "asOf"]) {
      __id
      totalCount
      edges {
        node {
          id
          ...TreatmentPlanListItem_treatmentPlan @arguments(asOf: $asOf)
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
  @refetchable(queryName: "RiskAnalysisPlansSectionRefetchQuery")
  @argumentDefinitions(
    filter: { type: "TreatmentPlanFilter", defaultValue: null }
    asOf: { type: "Datetime", defaultValue: null }
    includeUnplanned: { type: "Boolean!" }
    orderBy: { type: "TreatmentPlanOrder", defaultValue: null }
  ) {
    id
    createdAt
    matrixSize {
      rows
      cols
    }
    ...RiskAnalysisMatrices_analysis @arguments(asOf: $asOf)
    ...LinkedRiskListItem_analysis
    ...RiskAnalysisPlansSection_plans @arguments(filter: $filter, asOf: $asOf, orderBy: $orderBy)
    ...RiskAnalysisPlansSection_unplanned @include(if: $includeUnplanned)
  }
`;

type UnplannedPlans = {
  risks: ReadonlyArray<
    RiskAnalysisPlansSection_unplanned$data["scenarioRisks"]["edges"][number]["node"]
  >;
  totalCount: number;
  hasNext: boolean;
  isLoadingNext: boolean;
  loadNext: (count: number) => void;
  refetch: () => void;
};

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
  const { asOfDate, setAsOfDate } = useMatrixAsOf();
  const [isPending, startTransition] = useTransition();
  const [order, setOrder] = useState<Order>({ direction: "DESC", field: "" });
  const [analysis, refetch] = useRefetchableFragment<
    RiskAnalysisPlansSectionRefetchQuery,
    RiskAnalysisPlansSection_analysis$key
  >(riskAnalysisPlansSectionFragment, analysisKey);
  const { minDate, maxDate } = asOfDateBounds(analysis.createdAt);
  const clampedAsOfDate = clampDateInput(asOfDate, minDate, maxDate);
  const asOfNeedsClamp = clampedAsOfDate !== asOfDate;
  const skipInitialAsOfRefetch = useRef(!asOfNeedsClamp);
  const asOfPending = isPending || asOfNeedsClamp;
  const wantUnplanned = matrixAsOf(clampedAsOfDate) == null;
  const orderBy = useMemo(() => categoryOrder(order), [order]);
  const orderByRef = useRef(orderBy);
  const [unplannedReady, setUnplannedReady] = useState(!asOfNeedsClamp && wantUnplanned);
  if (!wantUnplanned && unplannedReady) {
    setUnplannedReady(false);
  }

  useEffect(() => {
    orderByRef.current = orderBy;
  });

  useEffect(() => {
    if (asOfNeedsClamp) {
      setAsOfDate(clampedAsOfDate, { clearCellFilter: false });
    }
  }, [asOfNeedsClamp, clampedAsOfDate, setAsOfDate]);

  useEffect(() => {
    if (skipInitialAsOfRefetch.current) {
      skipInitialAsOfRefetch.current = false;
      return;
    }

    let cancelled = false;
    startTransition(() => {
      refetch(
        {
          asOf: matrixAsOf(clampedAsOfDate),
          filter: treatmentPlanFilterFromCell(cell),
          includeUnplanned: wantUnplanned,
          orderBy: orderByRef.current,
        },
        {
          onComplete: (error) => {
            if (cancelled || error) {
              return;
            }

            setUnplannedReady(wantUnplanned);
          },
        },
      );
    });

    return () => {
      cancelled = true;
    };
  }, [clampedAsOfDate, cell, refetch, wantUnplanned]);

  if (!analysis.id || !analysis.matrixSize) {
    return null;
  }

  return (
    <div className="space-y-4">
      <RiskAnalysisMatrices
        analysisKey={analysis}
        asOfDate={clampedAsOfDate}
        isPending={asOfPending}
        onAsOfDateChange={setAsOfDate}
      />
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
        <RiskAnalysisPlansTable
          analysis={analysis}
          asOfDate={clampedAsOfDate}
          isPending={asOfPending}
          unplannedReady={unplannedReady}
          order={order}
          orderBy={orderBy}
          onOrderChange={setOrder}
        />
      </Suspense>
    </div>
  );
}

function RiskAnalysisPlansTable({
  analysis,
  asOfDate,
  isPending,
  unplannedReady,
  order,
  orderBy,
  onOrderChange,
}: {
  analysis: RiskAnalysisPlansSection_analysis$data;
  asOfDate: string;
  isPending: boolean;
  unplannedReady: boolean;
  order: Order;
  orderBy: ReturnType<typeof categoryOrder>;
  onOrderChange: (order: Order) => void;
}) {
  const { t } = useTranslation();
  const { cell } = useMatrixCellFilter();
  const [, startTransition] = useTransition();
  const asOf = matrixAsOf(asOfDate);
  const hasAsOf = asOf != null;
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
  const filter = useMemo(
    () => treatmentPlanFilterFromCell(cell),
    [cell],
  );
  const plansQueryRef = useRef({ asOf, filter });
  const skipFirstPlansRefetch = useRef(true);

  useEffect(() => {
    plansQueryRef.current = { asOf, filter };
  });

  useEffect(() => {
    if (skipFirstPlansRefetch.current) {
      skipFirstPlansRefetch.current = false;
      return;
    }
    startTransition(() => {
      refetchPlans({
        asOf: plansQueryRef.current.asOf,
        filter: plansQueryRef.current.filter,
        orderBy,
      });
    });
  }, [orderBy, refetchPlans]);

  if (!analysis.matrixSize) {
    return null;
  }

  const matrixSize = { rows: analysis.matrixSize.rows, cols: analysis.matrixSize.cols };
  const plans = plansData.treatmentPlans?.edges.map(edge => edge.node) ?? [];
  const connectionId = plansData.treatmentPlans?.__id ?? "";

  const renderTable = (unplanned: UnplannedPlans | null) => {
    const visibleUnplannedRisks = hasAsOf || cell ? [] : (unplanned?.risks ?? []);
    const unplannedTotal = unplanned?.totalCount ?? 0;
    const showMore = hasAsOf || cell
      ? plansHasNext
      : plansHasNext || (unplanned?.hasNext ?? false);
    const loadingMore = plansLoadingNext || (unplanned?.isLoadingNext ?? false);
    const reload = () => {
      startTransition(() => {
        refetchPlans({ filter, orderBy, asOf }, { fetchPolicy: "network-only" });
        unplanned?.refetch();
      });
    };
    const isEmpty = plans.length === 0 && visibleUnplannedRisks.length === 0;
    const rows = [
      ...plans.map(plan => ({ kind: "plan" as const, plan })),
      ...visibleUnplannedRisks.map(risk => ({ kind: "unplanned" as const, risk })),
    ];

    return (
      <div
        aria-busy={isPending}
        className={`space-y-4 transition-opacity ${isPending ? "opacity-60" : ""}`}
      >
        {unplannedTotal > 0 && (
          <div className="flex items-center gap-3 rounded-lg bg-warning px-4 py-3 text-sm text-txt-warning">
            <IconWarning size={16} className="shrink-0" />
            {t("riskAnalysisTreatmentPlansPage.unplannedWarning", {
              count: unplannedTotal,
            })}
          </div>
        )}
        <SortableContext value={{ order, changeOrder: onOrderChange }}>
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
                        readOnly={hasAsOf}
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
              unplanned?.loadNext(TREATMENT_PLAN_PAGE_SIZE);
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
  };

  if (hasAsOf || !unplannedReady) {
    return renderTable(null);
  }

  return (
    <RiskAnalysisUnplannedPlans
      analysisKey={analysis}
      orderBy={orderBy}
    >
      {unplanned => renderTable(unplanned)}
    </RiskAnalysisUnplannedPlans>
  );
}

function RiskAnalysisUnplannedPlans({
  analysisKey,
  orderBy,
  children,
}: {
  analysisKey: RiskAnalysisPlansSection_unplanned$key;
  orderBy: ReturnType<typeof categoryOrder>;
  children: (unplanned: UnplannedPlans) => ReactNode;
}) {
  const [, startTransition] = useTransition();
  const {
    data,
    hasNext,
    isLoadingNext,
    loadNext,
    refetch,
  } = usePaginationFragment<
    RiskAnalysisPlansSectionUnplannedQuery,
    RiskAnalysisPlansSection_unplanned$key
  >(riskAnalysisPlansSectionUnplannedFragment, analysisKey);
  const skipFirstRefetch = useRef(true);

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }
    startTransition(() => {
      refetch({ orderBy });
    });
  }, [orderBy, refetch, startTransition]);

  const risks = data.scenarioRisks?.edges.map(edge => edge.node) ?? [];

  return children({
    risks,
    totalCount: data.scenarioRisks?.totalCount ?? 0,
    hasNext,
    isLoadingNext,
    loadNext,
    refetch: () => {
      refetch({ orderBy }, { fetchPolicy: "network-only" });
    },
  });
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
