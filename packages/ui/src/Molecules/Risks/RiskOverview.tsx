// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
