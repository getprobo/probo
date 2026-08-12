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

export const riskAnalysisMatrixSizeOptions = [
  { key: "3x3", rows: 3, cols: 3 },
  { key: "4x4", rows: 4, cols: 4 },
  { key: "5x5", rows: 5, cols: 5 },
] as const;

export type RiskAnalysisMatrixSizeOption = (typeof riskAnalysisMatrixSizeOptions)[number]["key"];

export type MatrixSize = {
  rows: number;
  cols: number;
};

export function matrixSizeKey(size: MatrixSize): RiskAnalysisMatrixSizeOption {
  return `${size.rows}x${size.cols}` as RiskAnalysisMatrixSizeOption;
}

export function matrixSizeFromOption(key: RiskAnalysisMatrixSizeOption): MatrixSize {
  const option = riskAnalysisMatrixSizeOptions.find(size => size.key === key);
  if (!option) {
    throw new Error(`Invalid matrix size "${key}"`);
  }

  return { rows: option.rows, cols: option.cols };
}

export function formatMatrixSize(rows: number, cols: number): string {
  return `${rows}×${cols}`;
}
