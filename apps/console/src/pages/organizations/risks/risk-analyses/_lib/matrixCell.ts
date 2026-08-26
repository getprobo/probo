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

export const graphqlScoreTypes = {
  inherent: "INHERENT",
  net: "NET",
  residual: "RESIDUAL",
} as const;

export type GraphQLScoreType = (typeof graphqlScoreTypes)[MatrixScoreType];

export type MatrixCell = {
  type: MatrixScoreType;
  likelihood: number;
  impact: number;
};

export type TreatmentPlanFilterInput = {
  scoreType: GraphQLScoreType;
  likelihood: number;
  impact: number;
};

export function isMatrixScoreType(value: string | null): value is MatrixScoreType {
  return matrixScoreTypes.some(type => type === value);
}

export function sameMatrixCell(a: MatrixCell, b: MatrixCell): boolean {
  return a.type === b.type
    && a.likelihood === b.likelihood
    && a.impact === b.impact;
}

function parseAxis(value: string | null): number | null {
  if (value == null || value === "") {
    return null;
  }

  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed < 1) {
    return null;
  }

  return parsed;
}

export function parseMatrixCell(params: URLSearchParams): MatrixCell | null {
  const type = params.get("score");
  const likelihood = parseAxis(params.get("likelihood"));
  const impact = parseAxis(params.get("impact"));
  if (!isMatrixScoreType(type) || likelihood == null || impact == null) {
    return null;
  }

  return { type, likelihood, impact };
}

export function treatmentPlanFilterFromCell(
  cell: MatrixCell | null,
): TreatmentPlanFilterInput | null {
  if (cell == null) {
    return null;
  }

  return {
    scoreType: graphqlScoreTypes[cell.type],
    likelihood: cell.likelihood,
    impact: cell.impact,
  };
}
