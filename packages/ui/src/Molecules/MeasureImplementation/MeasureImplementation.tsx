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

import { measureStates } from "@probo/helpers";
import { clsx } from "clsx";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";

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
  UNKNOWN: "bg-highlight",
  NOT_IMPLEMENTED: "bg-border-danger",
};

export function MeasureImplementation({ measures, className }: Props) {
  const { t } = useTranslation();
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
        {t("ui.measureImplementation.title")}
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
            {t("ui.measureImplementation.complete")}
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
            {t(`ui.measureState.${state.toLowerCase()}`)}
          </div>
        ))}
      </div>
    </div>
  );
}
