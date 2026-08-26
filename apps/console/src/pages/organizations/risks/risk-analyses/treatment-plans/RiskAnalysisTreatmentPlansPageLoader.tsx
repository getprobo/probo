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

import { Suspense, useEffect, useMemo } from "react";
import { useQueryLoader } from "react-relay";
import { useParams, useSearchParams } from "react-router";

import type { RiskAnalysisTreatmentPlansPageQuery } from "#/__generated__/core/RiskAnalysisTreatmentPlansPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";

import { matrixAsOf, parseAsOfDate } from "../_lib/matrixAsOf";
import {
  parseMatrixCell,
  treatmentPlanFilterFromCell,
  type TreatmentPlanFilterInput,
} from "../_lib/matrixCell";

import RiskAnalysisTreatmentPlansPage, {
  riskAnalysisTreatmentPlansPageQuery,
} from "./RiskAnalysisTreatmentPlansPage";

export default function RiskAnalysisTreatmentPlansPageLoader() {
  const { riskAnalysisId } = useParams<{ riskAnalysisId: string }>();
  const [params] = useSearchParams();
  const search = params.toString();
  const [queryRef, loadQuery]
    = useQueryLoader<RiskAnalysisTreatmentPlansPageQuery>(riskAnalysisTreatmentPlansPageQuery);
  const { asOf, filter, includeUnplanned } = useMemo(() => {
    const parsed = new URLSearchParams(search);
    const nextAsOf = matrixAsOf(parseAsOfDate(parsed));

    return {
      asOf: nextAsOf,
      filter: treatmentPlanFilterFromCell(parseMatrixCell(parsed)),
      includeUnplanned: nextAsOf == null,
    };
  }, [search]);

  useEffect(() => {
    if (riskAnalysisId) {
      loadQuery({
        riskAnalysisId,
        filter,
        asOf,
        includeUnplanned,
      });
    }
  }, [asOf, filter, includeUnplanned, loadQuery, riskAnalysisId]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.riskAnalysisId === riskAnalysisId
    && queryRef.variables.asOf === asOf
    && queryRef.variables.includeUnplanned === includeUnplanned
    && sameTreatmentPlanFilter(queryRef.variables.filter, filter)
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <LinkCardSkeleton />;
  }

  return (
    <Suspense key={riskAnalysisId} fallback={<LinkCardSkeleton />}>
      <RiskAnalysisTreatmentPlansPage queryRef={currentQueryRef} />
    </Suspense>
  );
}

function sameTreatmentPlanFilter(
  loaded: RiskAnalysisTreatmentPlansPageQuery["variables"]["filter"],
  wanted: TreatmentPlanFilterInput | null,
): boolean {
  if (wanted == null) {
    return loaded == null;
  }

  return loaded != null
    && loaded.scoreType === wanted.scoreType
    && loaded.likelihood === wanted.likelihood
    && loaded.impact === wanted.impact;
}
