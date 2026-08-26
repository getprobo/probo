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

import { getRiskImpacts, getRiskLikelihoods } from "@probo/helpers";
import { Card, RisksChart } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { RiskAnalysisMatrices_analysis$key } from "#/__generated__/core/RiskAnalysisMatrices_analysis.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  graphqlScoreTypes,
  type MatrixCell,
  type MatrixScoreType,
} from "../_lib/matrixCell";
import { useMatrixCellFilter } from "../_lib/useMatrixCellFilter";

const severityColors = ["bg-txt-success", "bg-txt-warning", "bg-txt-danger"] as const;

export const riskAnalysisMatricesFragment = graphql`
  fragment RiskAnalysisMatrices_analysis on RiskAnalysis {
    matrixSize {
      rows
      cols
    }
    matrixCells {
      type
      likelihood
      impact
      count
    }
  }
`;

interface RiskAnalysisMatricesProps {
  analysisKey: RiskAnalysisMatrices_analysis$key;
}

function selectedChartCell(
  cell: MatrixCell | null,
  type: MatrixScoreType,
): { likelihood: number; impact: number } | null {
  if (cell == null || cell.type !== type) {
    return null;
  }

  return { likelihood: cell.likelihood, impact: cell.impact };
}

function cellCountsForScoreType(
  counts: ReadonlyArray<{
    type: string;
    likelihood: number;
    impact: number;
    count: number;
  }>,
  type: MatrixScoreType,
) {
  const graphqlType = graphqlScoreTypes[type];
  return counts
    .filter(cell => cell.type === graphqlType)
    .map(cell => ({
      likelihood: cell.likelihood,
      impact: cell.impact,
      count: cell.count,
    }));
}

export function RiskAnalysisMatrices({ analysisKey }: RiskAnalysisMatricesProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { cell, toggleCell } = useMatrixCellFilter();
  const analysis = useFragment(riskAnalysisMatricesFragment, analysisKey);
  if (!analysis.matrixSize) {
    return null;
  }

  const matrixSize = { rows: analysis.matrixSize.rows, cols: analysis.matrixSize.cols };
  const counts = analysis.matrixCells ?? [];
  const impacts = getRiskImpacts(t, matrixSize.cols);
  const likelihoods = getRiskLikelihoods(t, matrixSize.rows);
  const formatScale = (items: ReadonlyArray<{ value: number; label: string }>) =>
    items
      .map(item => t("riskAnalysisTreatmentPlansPage.legend.level", {
        value: item.value,
        label: item.label,
      }))
      .join(" · ");
  const severities = [
    t("ui.risk.severity.low"),
    t("ui.risk.severity.high"),
    t("ui.risk.severity.critical"),
  ];

  return (
    <Card padded className="space-y-6 overflow-visible">
      <div className="grid grid-cols-3 gap-6">
        <RisksChart
          organizationId={organizationId}
          type="inherent"
          cellCounts={cellCountsForScoreType(counts, "inherent")}
          matrixSize={matrixSize}
          variant="bare"
          selectedCell={selectedChartCell(cell, "inherent")}
          onCellSelect={next => toggleCell({ type: "inherent", ...next })}
        />
        <RisksChart
          organizationId={organizationId}
          type="net"
          cellCounts={cellCountsForScoreType(counts, "net")}
          matrixSize={matrixSize}
          variant="bare"
          selectedCell={selectedChartCell(cell, "net")}
          onCellSelect={next => toggleCell({ type: "net", ...next })}
        />
        <RisksChart
          organizationId={organizationId}
          type="residual"
          cellCounts={cellCountsForScoreType(counts, "residual")}
          matrixSize={matrixSize}
          variant="bare"
          selectedCell={selectedChartCell(cell, "residual")}
          onCellSelect={next => toggleCell({ type: "residual", ...next })}
        />
      </div>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-1 text-xs text-txt-secondary">
          <div>
            {t("riskAnalysisTreatmentPlansPage.legend.impact", {
              levels: formatScale(impacts),
            })}
          </div>
          <div>
            {t("riskAnalysisTreatmentPlansPage.legend.likelihood", {
              levels: formatScale(likelihoods),
            })}
          </div>
        </div>
        <div className="flex gap-3">
          {severities.map((label, index) => (
            <div key={label} className="flex items-center gap-1 text-xs text-txt-secondary">
              <div className={`size-[10px] rounded-xs ${severityColors[index]}`} />
              <span>{label}</span>
            </div>
          ))}
        </div>
      </div>
    </Card>
  );
}
