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

import { Input } from "@probo/ui";
import { useTranslation } from "react-i18next";

import type { ScoreLevelDraft } from "./scoreLevels";

interface ScoreLevelsFieldProps {
  id: string;
  label: string;
  value: ScoreLevelDraft[];
  maxScore: number;
  error?: string;
  onChange: (value: ScoreLevelDraft[]) => void;
}

export function ScoreLevelsField({
  id,
  label,
  value,
  maxScore,
  error,
  onChange,
}: ScoreLevelsFieldProps) {
  const { t } = useTranslation();

  const update = (index: number, patch: Partial<ScoreLevelDraft>) => {
    onChange(value.map((level, levelIndex) => (
      levelIndex === index ? { ...level, ...patch } : level
    )));
  };

  return (
    <fieldset className="space-y-2" aria-describedby={error ? `${id}-error` : undefined}>
      <legend className="text-sm font-medium text-txt-primary">{label}</legend>
      <div className="grid grid-cols-2 gap-3">
        {value.map((level, index) => {
          const levelName = t(`ui.risk.severity.${level.level.toLowerCase()}`);
          const checkboxId = `${id}-${level.level}`;
          const scoreId = `${id}-${level.level}-min`;
          return (
            <div key={level.level} className="space-y-2 rounded-xl border border-border-low p-3">
              <div className="flex items-center gap-2">
                <input
                  id={checkboxId}
                  type="checkbox"
                  className="size-4"
                  checked={level.included}
                  onChange={event => update(index, { included: event.target.checked })}
                />
                <label htmlFor={checkboxId} className="text-sm font-medium text-txt-primary">
                  {levelName}
                </label>
              </div>
              <div className={`flex items-center justify-between gap-2 ${level.included ? "" : "opacity-50"}`}>
                <label htmlFor={scoreId} className="text-sm text-txt-secondary">
                  {t("scoreLevels.minimum")}
                </label>
                <Input
                  id={scoreId}
                  type="number"
                  min={1}
                  max={maxScore}
                  className="w-20"
                  disabled={!level.included}
                  aria-label={t("scoreLevels.minScore", { level: levelName })}
                  value={Number.isNaN(level.minScore) ? "" : level.minScore}
                  onChange={event => update(index, {
                    minScore: event.target.value === "" ? Number.NaN : Number(event.target.value),
                  })}
                />
              </div>
            </div>
          );
        })}
      </div>
      {error && (
        <p id={`${id}-error`} className="text-sm text-txt-danger">
          {error}
        </p>
      )}
    </fieldset>
  );
}
