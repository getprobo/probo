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

import type { INodeProperties } from 'n8n-workflow';

export type ScoreLevelInput = {
	level: string;
	minScore: number;
};

// Same ratio bands as riskmanagement.DefaultScoreLevels: medium at 5/25 of the
// matrix maximum, high at 10/25, and critical at 15/25, each rounded up.
export function defaultScoreLevels(rows: number, cols: number): ScoreLevelInput[] {
	const maxScore = rows * cols;
	const minScore = (numerator: number) => Math.trunc((maxScore * numerator + 24) / 25);

	return [
		{ level: 'LOW', minScore: 1 },
		{ level: 'MEDIUM', minScore: minScore(5) },
		{ level: 'HIGH', minScore: minScore(10) },
		{ level: 'CRITICAL', minScore: minScore(15) },
	];
}

const fiveByFiveScoreLevels = defaultScoreLevels(5, 5);

function sameScoreLevels(left: ScoreLevelInput[], right: ScoreLevelInput[]): boolean {
	return left.length === right.length
		&& left.every((item, index) => item.level === right[index].level && item.minScore === right[index].minScore);
}

// The create field is prefilled per matrix size. A stored 5×5 preset still
// has to follow the matrix the user selected, because that preset is invalid
// on a 3×3 matrix and is not the ratio mapping for 4×4.
export function resolveCreateScoreLevels(value: unknown, rows: number, cols: number): ScoreLevelInput[] {
	const selected = defaultScoreLevels(rows, cols);
	if (value == null || value === '') {
		return selected;
	}

	const parsed = parseScoreLevels(value);
	if ((rows !== 5 || cols !== 5) && sameScoreLevels(parsed, fiveByFiveScoreLevels)) {
		return selected;
	}

	return parsed;
}

export const scoreLevelsSelection = `
	scoreLevels {
		level
		minScore
	}
`;

export function parseScoreLevels(value: unknown): ScoreLevelInput[] {
	let parsed = value;
	if (typeof value === 'string') {
		parsed = JSON.parse(value);
	}

	if (!Array.isArray(parsed)) {
		throw new Error('Score levels must be a JSON array');
	}

	return parsed.map((item) => {
		if (item == null || typeof item !== 'object') {
			throw new Error('Each score level must be an object with level and minScore');
		}

		const record = item as { level?: unknown; minScore?: unknown };
		if (typeof record.level !== 'string' || typeof record.minScore !== 'number') {
			throw new Error('Each score level must be an object with level and minScore');
		}

		return { level: record.level, minScore: record.minScore };
	});
}

export function scoreLevelsProperty(matrixSize: string, rows: number, cols: number): INodeProperties {
	return {
		displayName: 'Score Levels',
		name: 'scoreLevels',
		type: 'json',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['create'],
				matrixSize: [matrixSize],
			},
		},
		default: JSON.stringify(defaultScoreLevels(rows, cols)),
		description: 'Score-to-level mapping for this matrix. Each item has level (LOW, MEDIUM, HIGH, or CRITICAL, in that order) and minScore. Include at least two levels. The first minScore must be 1. Prefilled values are the ratio bands for the selected matrix size.',
		required: true,
	};
}
