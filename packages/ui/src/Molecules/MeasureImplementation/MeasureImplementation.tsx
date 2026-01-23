import { getMeasureStateLabel, measureStates } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { clsx } from "clsx";
import type { ComponentProps } from "react";

import type { MeasureBadge } from "../Badge/MeasureBadge";

type MeasureState = ComponentProps<typeof MeasureBadge>["state"];

type Props = {
  measures: { state: MeasureState }[];
  className?: string;
};

const stateToColor: Record<MeasureState, string> = {
  IMPLEMENTED: "bg-border-success",
  IN_PROGRESS: "bg-border-warning",
  NOT_APPLICABLE: "bg-border-info",
  NOT_STARTED: "bg-highlight",
};

export function MeasureImplementation({ measures, className }: Props) {
  const { __ } = useTranslate();
  const counts = measures.reduce(
    (acc, measure) => {
      acc[measure.state] = (acc[measure.state] ?? 0) + 1;
      return acc;
    },
    {} as Record<MeasureState, number>,
  );
  const percent = Math.round(
    (100
      * ((counts["IMPLEMENTED"] ?? 0) + (counts["NOT_APPLICABLE"] ?? 0)))
    / measures.length,
  );
  return (
    <div className={clsx("space-y-3", className)}>
      <h2 className="text-base font-medium">
        {__("Measure implementation")}
      </h2>
      <div className="h-2 rounded overflow-hidden bg-highlight flex justify-stretch item-stretch">
        {measureStates.map(state => (
          <div
            key={state}
            className={clsx(stateToColor[state])}
            style={{
              flexGrow: counts[state] ?? 0,
            }}
          />
        ))}
      </div>
      <div className="flex gap-4 text-sm">
        {!isNaN(percent) && (
          <div className="mr-auto">
            {percent}
            %
            {__("Complete")}
          </div>
        )}
        {measureStates.map(state => (
          <div
            key={state}
            className="text-sm text-txt-secondary flex items-center gap-[6px]"
          >
            <div
              className={clsx(
                "size-[10px] rounded-full",
                stateToColor[state],
              )}
            >
            </div>
            {getMeasureStateLabel(__, state)}
          </div>
        ))}
      </div>
    </div>
  );
}
