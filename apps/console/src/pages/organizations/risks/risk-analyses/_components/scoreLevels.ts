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

import {
  defaultScoreLevels,
  type RiskScoreLevel,
  type RiskScoreLevelMapping,
  riskScoreLevelOrder,
} from "@probo/helpers";

export type ScoreLevelDraft = {
  level: RiskScoreLevel;
  included: boolean;
  minScore: number;
};

export type ScoreLevelError = "atLeastTwo" | "startAtOne" | "increasing" | "withinMatrix";

export function draftScoreLevels(rows: number, cols: number): ScoreLevelDraft[] {
  return defaultScoreLevels(rows, cols).map(level => ({
    level: level.level,
    included: true,
    minScore: level.minScore,
  }));
}

export function draftFromScoreLevels(
  levels: ReadonlyArray<RiskScoreLevelMapping>,
  rows: number,
  cols: number,
): ScoreLevelDraft[] {
  const configured = new Map(levels.map(level => [level.level, level.minScore]));
  const defaults = defaultScoreLevels(rows, cols);

  return riskScoreLevelOrder.map(level => ({
    level,
    included: configured.has(level),
    minScore: configured.get(level) ?? defaults.find(item => item.level === level)?.minScore ?? 1,
  }));
}

export function scoreLevelsFromDraft(
  draft: ReadonlyArray<ScoreLevelDraft>,
): RiskScoreLevelMapping[] {
  return draft
    .filter(level => level.included)
    .map(level => ({
      level: level.level,
      minScore: Number(level.minScore),
    }));
}

export function scoreLevelDraftError(
  draft: ReadonlyArray<ScoreLevelDraft>,
  maxScore: number,
): ScoreLevelError | null {
  const included = draft.filter(level => level.included);
  if (included.length < 2) {
    return "atLeastTwo";
  }

  let previous = 0;
  for (const [index, level] of included.entries()) {
    const minScore = Number(level.minScore);
    if (!Number.isInteger(minScore) || minScore < 1 || minScore > maxScore) {
      return "withinMatrix";
    }

    if (index === 0 && minScore !== 1) {
      return "startAtOne";
    }

    if (index > 0 && minScore <= previous) {
      return "increasing";
    }

    previous = minScore;
  }

  return null;
}

export function scoreLevelRanges(
  levels: ReadonlyArray<RiskScoreLevelMapping>,
  maxScore: number,
): { level: RiskScoreLevel; range: string }[] {
  return levels.map((level, index) => {
    const next = levels[index + 1];
    const upper = next ? next.minScore - 1 : maxScore;

    return { level: level.level, range: `${level.minScore}-${upper}` };
  });
}

export function formatScoreLevelSummary(
  levels: ReadonlyArray<RiskScoreLevelMapping>,
  maxScore: number,
  label: (level: RiskScoreLevel) => string,
): string {
  return levels.map((level, index) => {
    const next = levels[index + 1];
    const upper = next ? next.minScore - 1 : maxScore;
    const first = index === 0;
    const last = next == null;
    let band = `${level.minScore}–${upper}`;
    if (first && level.minScore <= 1) {
      band = `≤${upper}`;
    } else if (last) {
      band = `≥${level.minScore}`;
    } else if (level.minScore === upper) {
      band = String(level.minScore);
    }

    return `${label(level.level)} ${band}`;
  }).join(", ");
}
