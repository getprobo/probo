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

import { getRiskImpacts, getRiskLikelihoods, getRiskScoreLevel, groupBy } from "@probo/helpers";
import { clsx } from "clsx";
import { Fragment, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import { Card } from "../../Atoms/Card/Card";
import {
  Dropdown,
  DropdownItem,
  DropdownSeparator,
} from "../../Atoms/Dropdown/Dropdown";
import { IconChevronRight, IconFire3 } from "../../Atoms/Icons";

import { levelColors } from "./constants";

type ChartType = "inherent" | "residual" | "net";

type Props = {
  organizationId: string;
  type: ChartType;
  risks?: ReadonlyArray<Risk>;
  matrixSize?: { rows: number; cols: number };
  variant?: "default" | "bare";
  selectedCell?: { likelihood: number; impact: number } | null;
  onCellSelect?: (cell: { likelihood: number; impact: number }) => void;
};

type Risk = {
  id: string;
  name: string;
  inherentLikelihood?: number;
  inherentImpact?: number;
  residualLikelihood?: number;
  residualImpact?: number;
  netLikelihood?: number;
  netImpact?: number;
};

const cellKey = (impact: number, likelihood: number) =>
  `${impact}-${likelihood}`;

function inMatrix(score: number, max: number): boolean {
  return Number.isInteger(score) && score >= 1 && score <= max;
}

function cellScores(risk: Risk, type: ChartType): { impact: number; likelihood: number } | null {
  if (type === "net") {
    if (risk.netImpact == null || risk.netLikelihood == null) {
      return null;
    }
    return { impact: risk.netImpact, likelihood: risk.netLikelihood };
  }
  if (type === "residual") {
    if (risk.residualImpact == null || risk.residualLikelihood == null) {
      return null;
    }
    return { impact: risk.residualImpact, likelihood: risk.residualLikelihood };
  }
  if (risk.inherentImpact == null || risk.inherentLikelihood == null) {
    return null;
  }
  return { impact: risk.inherentImpact, likelihood: risk.inherentLikelihood };
}

function chartTitle(type: ChartType, t: (key: string) => string): string {
  if (type === "inherent") {
    return t("ui.risk.initial");
  }
  if (type === "net") {
    return t("ui.risk.net");
  }
  return t("ui.risk.residual");
}

/**
 * Displays a grid of risk grouped by impact & likelihood
 */
export function RisksChart({
  organizationId,
  type,
  risks,
  matrixSize,
  variant = "default",
  selectedCell,
  onCellSelect,
}: Props) {
  const { t } = useTranslation();
  const likelihoodMax = matrixSize?.rows ?? 5;
  const impactMax = matrixSize?.cols ?? 5;
  const bare = variant === "bare";

  const legend = [
    t("ui.risk.severity.low"),
    t("ui.risk.severity.high"),
    t("ui.risk.severity.critical"),
  ];

  const impacts = getRiskImpacts(t, impactMax).reverse();
  const likelihoods = getRiskLikelihoods(t, likelihoodMax);
  const labelColumn = bare ? "1.25rem" : "90px";

  const riskMap = useMemo(() => {
    const placed = [];
    for (const risk of risks ?? []) {
      const scores = cellScores(risk, type);
      if (
        scores == null
        || !inMatrix(scores.impact, impactMax)
        || !inMatrix(scores.likelihood, likelihoodMax)
      ) {
        continue;
      }
      placed.push({
        risk,
        key: cellKey(scores.impact, scores.likelihood),
      });
    }
    const grouped = groupBy(placed, item => item.key);
    return Object.fromEntries(
      Object.entries(grouped).map(([key, items]) => [key, items.map(item => item.risk)]),
    );
  }, [impactMax, likelihoodMax, risks, type]);

  const grid = (
    <div className="min-w-0 flex-1">
      <div
        className="grid gap-1 w-full"
        style={{
          gridTemplateColumns: `${labelColumn} repeat(${likelihoods.length}, minmax(0, 1fr))`,
        }}
      >
        {impacts.map(impact => (
          <Fragment key={impact.value}>
            <div className="pr-1 text-right text-xs text-txt-secondary flex items-center justify-end">
              {bare ? impact.value : `${impact.label} (${impact.value})`}
            </div>
            {likelihoods.map(likelihood => (
              <RisksChartCell
                key={likelihood.value}
                impact={impact.value}
                likelihood={likelihood.value}
                matrixSize={{ rows: likelihoodMax, cols: impactMax }}
                organizationId={organizationId}
                selected={
                  selectedCell?.likelihood === likelihood.value
                  && selectedCell.impact === impact.value
                }
                risks={
                  riskMap[
                    cellKey(
                      impact.value,
                      likelihood.value,
                    )
                  ]
                }
                onSelect={onCellSelect
                  ? () => onCellSelect({
                      likelihood: likelihood.value,
                      impact: impact.value,
                    })
                  : undefined}
              />
            ))}
          </Fragment>
        ))}
        <div></div>
        {likelihoods.map(likelihood => (
          <div
            className="text-center text-xs text-txt-secondary mt-2"
            key={likelihood.value}
          >
            {bare ? likelihood.value : `${likelihood.label} (${likelihood.value})`}
          </div>
        ))}
      </div>
      <div
        className="mt-2 text-center text-xs font-medium whitespace-nowrap"
        style={{ paddingLeft: labelColumn }}
      >
        {t("ui.risk.likelihood.label")}
      </div>
    </div>
  );

  if (bare) {
    return (
      <div className="overflow-visible text-txt-primary">
        <h2 className="mb-3 text-sm font-semibold">
          {chartTitle(type, t)}
        </h2>
        <div className="flex gap-3 overflow-visible">
          <div
            className="shrink-0 self-stretch text-xs font-medium text-center whitespace-nowrap"
            style={{ writingMode: "sideways-lr" }}
          >
            {t("ui.risk.impact.label")}
          </div>
          {grid}
        </div>
      </div>
    );
  }

  return (
    <Card padded className="text-txt-primary">
      <div className="flex justify-between items-center mb-6">
        <h2 className="font-semibold text-lg">
          {chartTitle(type, t)}
        </h2>
        <div className="flex gap-3">
          {legend.map((label, i) => (
            <div
              key={label}
              className="flex items-center gap-1 text-xs"
            >
              <div
                className={clsx(
                  "size-[10px] rounded-xs",
                  levelColors[i].color,
                )}
              />
              <span>{label}</span>
            </div>
          ))}
        </div>
      </div>
      <div className="flex gap-6">
        <div
          className="text-xs font-medium flex-none text-center"
          style={{ writingMode: "sideways-lr" }}
        >
          {t("ui.risk.impact.label")}
        </div>
        {grid}
      </div>
    </Card>
  );
}

function RisksChartCell({
  risks,
  impact,
  likelihood,
  matrixSize,
  organizationId,
  selected = false,
  onSelect,
}: {
  risks?: ReadonlyArray<Risk>;
  impact: number;
  likelihood: number;
  matrixSize: { rows: number; cols: number };
  organizationId: string;
  selected?: boolean;
  onSelect?: () => void;
}) {
  const { t } = useTranslation();
  const level = getRiskScoreLevel(impact * likelihood, matrixSize);
  const baseClass
    = "flex items-center justify-center aspect-square rounded-xl text-txt-invert text-sm font-semibold";
  const selectedClass = selected
    ? "outline-2 outline-offset-2 outline-txt-primary"
    : undefined;

  if (onSelect) {
    return (
      <button
        type="button"
        aria-pressed={selected}
        aria-label={t("ui.risk.cell", {
          count: risks?.length ?? 0,
          likelihood,
          impact,
        })}
        className={clsx(
          baseClass,
          risks ? levelColors[level].color : levelColors[level].bg,
          "cursor-pointer",
          selectedClass,
        )}
        onClick={onSelect}
      >
        {risks?.length ?? ""}
      </button>
    );
  }

  if (!risks) {
    return <div className={clsx(baseClass, levelColors[level].bg)}></div>;
  }

  const infos = [
    { label: t("ui.risk.numberOfRisks"), value: risks.length },
    { label: t("ui.risk.impact.label"), value: impact },
    { label: t("ui.risk.likelihood.label"), value: likelihood },
  ];

  return (
    <Dropdown
      className="text-sm w-75 p-4 space-y-1"
      toggle={(
        <button
          className={clsx(
            baseClass,
            levelColors[level].color,
            "cursor-pointer",
          )}
        >
          {risks.length}
        </button>
      )}
    >
      {infos.map(info => (
        <div
          key={info.label}
          className="flex items-center justify-between gap-4"
        >
          <div className="text-txt-secondary">{info.label}</div>
          <div className="text-txt-primary">{info.value}</div>
        </div>
      ))}
      <DropdownSeparator className="my-3" />
      <div className="flex items-center justify-between gap-4">
        <div className="text-txt-secondary">{t("ui.risk.score")}</div>
        <div className="text-txt-primary">{impact * likelihood}</div>
      </div>
      <DropdownSeparator className="my-3" />
      <div className="text-txt-secondary mb-1">
        {t("ui.risk.linkedRisks")}
      </div>
      {risks.map(risk => (
        <DropdownItem key={risk.id} asChild>
          <Link
            to={`/organizations/${organizationId}/risk-management/risks/${risk.id}`}
          >
            <IconFire3 size={16} className="flex-none" />
            {risk.name}
            <IconChevronRight
              size={16}
              className="flex-none ml-auto"
            />
          </Link>
        </DropdownItem>
      ))}
    </Dropdown>
  );
}
