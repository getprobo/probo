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

import { getRiskImpacts, getRiskLikelihoods, groupBy } from "@probo/helpers";
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

type Props = {
  organizationId: string;
  type: "inherent" | "residual";
  risks?: Risk[];
};

type Risk = {
  id: string;
  name: string;
  inherentLikelihood: number;
  inherentImpact: number;
  residualLikelihood: number;
  residualImpact: number;
};

const getLevel = (score: number): 0 | 1 | 2 => {
  if (score >= 15) {
    return 2;
  }
  if (score >= 5) {
    return 1;
  }
  return 0;
};

const cellKey = (impact: number, likelihood: number) =>
  `${impact}-${likelihood}`;

/**
 * Displays a grid of risk grouped by impact & likelihood
 */
export function RisksChart({ organizationId, type, risks }: Props) {
  const { t } = useTranslation();

  const legend = [
    t("ui.risk.severity.low"),
    t("ui.risk.severity.high"),
    t("ui.risk.severity.critical"),
  ];

  const impacts = getRiskImpacts(t).reverse();
  const likelihoods = getRiskLikelihoods(t);
  const impactField
    = type === "inherent" ? "inherentImpact" : "residualImpact";
  const likelihoodField
    = type === "inherent" ? "inherentLikelihood" : "residualLikelihood";

  const riskMap = useMemo(() => {
    return groupBy(risks ?? [], risk =>
      cellKey(risk[impactField], risk[likelihoodField]),
    );
  }, [impactField, likelihoodField, risks]);

  return (
    <Card padded className="text-txt-primary">
      <div className="flex justify-between items-center mb-6">
        <h2 className="font-semibold text-lg">
          {type === "inherent"
            ? t("ui.risk.initial")
            : t("ui.risk.residual")}
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
      {/* Grid */}
      <div className="flex gap-6">
        <div
          className="text-xs font-medium flex-none text-center"
          style={{ writingMode: "sideways-lr" }}
        >
          {t("ui.risk.impact.label")}
        </div>
        <div className="grid grid-cols-[90px_1fr_1fr_1fr_1fr_1fr] gap-1 w-full">
          {impacts.map(impact => (
            <Fragment key={impact.value}>
              <div className="pr-2 text-right text-xs text-txt-secondary flex items-center">
                {impact.label}
                {" "}
                (
                {impact.value}
                )
              </div>
              {likelihoods.map(likelihood => (
                <RisksChartCell
                  key={likelihood.value}
                  impact={impact.value}
                  likelihood={likelihood.value}
                  organizationId={organizationId}
                  risks={
                    riskMap[
                      cellKey(
                        impact.value,
                        likelihood.value,
                      )
                    ]
                  }
                />
              ))}
            </Fragment>
          ))}
          {/* X axis */}
          <div></div>
          {likelihoods.map(likelihood => (
            <div
              className="text-center text-xs text-txt-secondary mt-4"
              key={likelihood.value}
            >
              {likelihood.label}
              {" "}
              (
              {likelihood.value}
              )
              {likelihood.value === 3 && (
                <div className="text-xs text-txt-primary font-medium flex-none text-center mt-3">
                  {t("ui.risk.likelihood.label")}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </Card>
  );
}

function RisksChartCell({
  risks,
  impact,
  likelihood,
  organizationId,
}: {
  risks?: Risk[];
  impact: number;
  likelihood: number;
  organizationId: string;
}) {
  const { t } = useTranslation();
  const level = getLevel(impact * likelihood);
  const baseClass
    = "flex items-center justify-center aspect-square rounded-xl text-txt-invert text-sm font-semibold";
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
        <div className="text-txt-secondary">Risk Score</div>
        <div className="text-txt-primary">{impact * likelihood}</div>
      </div>
      <DropdownSeparator className="my-3" />
      <div className="text-txt-secondary mb-1">
        {t("ui.risk.linkedRisks")}
      </div>
      {risks.map(risk => (
        <DropdownItem key={risk.id} asChild>
          <Link
            to={`/organizations/${organizationId}/risks/${risk.id}`}
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
