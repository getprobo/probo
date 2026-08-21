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

export const matrixScoreTypes = ["inherent", "net", "residual"] as const;

export type MatrixScoreType = (typeof matrixScoreTypes)[number];

export type MatrixCell = {
  type: MatrixScoreType;
  likelihood: number;
  impact: number;
};

export type MatrixScores = {
  inherentLikelihood?: number | null;
  inherentImpact?: number | null;
  residualLikelihood?: number | null;
  residualImpact?: number | null;
  netLikelihood?: number | null;
  netImpact?: number | null;
};

export function isMatrixScoreType(value: string | null): value is MatrixScoreType {
  return matrixScoreTypes.some(type => type === value);
}

export function sameMatrixCell(a: MatrixCell, b: MatrixCell): boolean {
  return a.type === b.type
    && a.likelihood === b.likelihood
    && a.impact === b.impact;
}

export function matchesMatrixCell(scores: MatrixScores, cell: MatrixCell): boolean {
  if (cell.type === "net") {
    return scores.netLikelihood === cell.likelihood
      && scores.netImpact === cell.impact;
  }

  if (cell.type === "residual") {
    return scores.residualLikelihood === cell.likelihood
      && scores.residualImpact === cell.impact;
  }

  return scores.inherentLikelihood === cell.likelihood
    && scores.inherentImpact === cell.impact;
}
