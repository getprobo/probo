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
	{ name: '3×3', value: '3x3' },
	{ name: '4×4', value: '4x4' },
	{ name: '5×5', value: '5x5' },
] as const;

const riskAnalysisMatrixSizes: Record<string, { rows: number; cols: number }> = {
	'3x3': { rows: 3, cols: 3 },
	'4x4': { rows: 4, cols: 4 },
	'5x5': { rows: 5, cols: 5 },
};

export function parseRiskAnalysisMatrixSize(value: string): { rows: number; cols: number } {
	if (!Object.prototype.hasOwnProperty.call(riskAnalysisMatrixSizes, value)) {
		throw new Error(`Invalid matrix size "${value}"; must be one of 3x3, 4x4, or 5x5`);
	}

	return riskAnalysisMatrixSizes[value];
}
