import { getRiskImpacts, getRiskLikelihoods, groupBy } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { clsx } from "clsx";
import { Fragment, useMemo } from "react";
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
  if (score > 6) {
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
  const { __ } = useTranslate();

  const legend = [__("Low"), __("Medium"), __("High")];

  const impacts = getRiskImpacts(__).reverse();
  const likelihoods = getRiskLikelihoods(__);
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
            ? __("Inherent Risk")
            : __("Residual Risk")}
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
          {__("Impact")}
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
                  {__("Likelihood")}
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
  const { __ } = useTranslate();
  const level = getLevel(impact * likelihood);
  const baseClass
    = "flex items-center justify-center aspect-square rounded-xl text-txt-invert text-sm font-semibold";
  if (!risks) {
    return <div className={clsx(baseClass, levelColors[level].bg)}></div>;
  }

  const infos = [
    { label: __("Number of risks"), value: risks.length },
    { label: __("Impact"), value: impact },
    { label: __("Likelihood"), value: likelihood },
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
      <div className="text-txt-secondary mb-1">{__("Linked Risks")}</div>
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
