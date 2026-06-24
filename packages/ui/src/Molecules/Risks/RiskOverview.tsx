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

import {
  getRiskImpacts,
  getRiskLikelihoods,
  getSeverity,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { clsx } from "clsx";

import { Card } from "../../Atoms/Card/Card";

import { levelColors } from "./constants";

type Props = {
  type: "inherent" | "residual";
  risk?: Risk;
};

type Risk = {
  inherentLikelihood: number;
  inherentImpact: number;
  residualLikelihood: number;
  residualImpact: number;
};

const getColor = (score: number): string => {
  const clamped = Math.min(5, Math.max(1, score));
  return levelColors[Math.ceil(clamped / 2) - 1].color;
};

export function RiskOverview({ type, risk }: Props) {
  const { __ } = useTranslate();
  const impact = risk?.[`${type}Impact`] ?? 0;
  const likelihood = risk?.[`${type}Likelihood`] ?? 0;
  const severity = getSeverity(__, impact * likelihood);
  return (
    <Card padded>
      <h2 className="font-semibold text-base mb-6">
        {type === "inherent" ? __("Initial Risk") : __("Residual Risk")}
      </h2>
      <div className="grid grid-cols-2 gap-4 mb-4">
        <RiskOverviewBadge
          label={__("Impact")}
          textCb={getRiskImpacts}
          score={impact}
        />
        <RiskOverviewBadge
          label={__("Likelihood")}
          textCb={getRiskLikelihoods}
          score={likelihood}
        />
      </div>
      <div className="space-y-2">
        <div className="font-medium text-xs">{__("Severity")}</div>
        <div
          className={clsx(
            severity?.bg,
            severity?.color,
            "py-2 text-sm font-semibold rounded-lg text-center",
          )}
        >
          {severity?.label}
        </div>
      </div>
    </Card>
  );
}

function RiskOverviewBadge({
  score,
  label,
  textCb,
}: {
  score: number;
  label: string;
  textCb: (t: (s: string) => string) => { value: number; label: string }[];
}) {
  const { __ } = useTranslate();
  return (
    <div className="space-y-2">
      <div className="font-medium text-xs">{__(label)}</div>
      <div
        className={clsx(
          getColor(score),
          "py-2 text-sm font-semibold rounded-lg text-txt-invert text-center",
        )}
      >
        {textCb(__).find(i => i.value === score)?.label}
        {" "}
        (
        {score}
        )
      </div>
    </div>
  );
}
