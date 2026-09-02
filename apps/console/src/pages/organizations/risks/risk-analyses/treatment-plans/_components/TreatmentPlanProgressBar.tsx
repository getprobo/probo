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

import { useTranslation } from "react-i18next";

interface TreatmentPlanProgressBarProps {
  done: number;
  inProgress: number;
  notImplemented: number;
  total: number;
}

export function TreatmentPlanProgressBar({
  done,
  inProgress,
  notImplemented,
  total,
}: TreatmentPlanProgressBarProps) {
  const { t } = useTranslation();
  const doneEnd = total === 0 ? 0 : Math.round((done / total) * 100);
  const inProgressEnd = total === 0
    ? 0
    : Math.round(((done + inProgress) / total) * 100);
  const paintedEnd = total === 0
    ? 0
    : Math.round(((done + inProgress + notImplemented) / total) * 100);
  const donePct = doneEnd;
  const inProgressPct = inProgressEnd - doneEnd;
  const notImplementedPct = paintedEnd - inProgressEnd;

  return (
    <div className="flex w-28 items-center gap-2">
      <div className="flex h-1.5 flex-1 overflow-hidden rounded-full bg-border-low">
        <div
          className="h-full bg-txt-success"
          style={{ width: `${donePct}%` }}
        />
        <div
          className="h-full bg-txt-warning"
          style={{ width: `${inProgressPct}%` }}
        />
        <div
          className="h-full bg-txt-danger"
          style={{ width: `${notImplementedPct}%` }}
        />
      </div>
      <span className="text-xs tabular-nums text-txt-secondary">
        {total === 0
          ? t("treatmentPlanListItem.progressEmpty")
          : t("treatmentPlanListItem.progress", {
              done,
              total,
            })}
      </span>
    </div>
  );
}
