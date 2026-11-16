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

type Translator = (s: string) => string;

type RiskScaleLevel = {
    value: number;
    label: string;
};

const IMPACT_5 = [
    "negligible",
    "low",
    "moderate",
    "significant",
    "catastrophic",
] as const;

const LIKELIHOOD_5 = [
    "improbable",
    "remote",
    "occasional",
    "probable",
    "frequent",
] as const;

function scaleLevels(
    t: Translator,
    prefix: string,
    size: number,
): RiskScaleLevel[] {
    return Array.from({ length: size }, (_, index) => {
        const value = index + 1;
        return {
            value,
            label: t(`${prefix}.${size}.${value}`),
        };
    });
}

export function getRiskImpacts(t: Translator, size = 5): RiskScaleLevel[] {
    if (size === 5) {
        return IMPACT_5.map((key, index) => ({
            value: index + 1,
            label: t(`helpers.riskImpact.${key}`),
        }));
    }

    return scaleLevels(t, "helpers.riskScales.impact", size);
}

export function getTreatment(t: Translator, treatment?: string): string {
    switch (treatment) {
        case "MITIGATED":
            return t("helpers.riskTreatment.mitigate");
        case "ACCEPTED":
            return t("helpers.riskTreatment.accept");
        case "TRANSFERRED":
            return t("helpers.riskTreatment.transfer");
        case "AVOIDED":
            return t("helpers.riskTreatment.avoid");
        default:
            return t("helpers.common.unknown");
    }
}

export function getRiskLikelihoods(t: Translator, size = 5): RiskScaleLevel[] {
    if (size === 5) {
        return LIKELIHOOD_5.map((key, index) => ({
            value: index + 1,
            label: t(`helpers.riskLikelihood.${key}`),
        }));
    }

    return scaleLevels(t, "helpers.riskScales.likelihood", size);
}

export type RiskMatrixSize = {
    rows: number;
    cols: number;
};

export const riskScoreLevelOrder = ["LOW", "MEDIUM", "HIGH", "CRITICAL"] as const;

export type RiskScoreLevel = (typeof riskScoreLevelOrder)[number];

export type RiskScoreLevelMapping = {
    level: RiskScoreLevel;
    minScore: number;
};

const LEVEL_INDEX: Record<RiskScoreLevel, 0 | 1 | 2 | 3> = {
    LOW: 0,
    MEDIUM: 1,
    HIGH: 2,
    CRITICAL: 3,
};

const DEFAULT_MATRIX_SIZE: RiskMatrixSize = { rows: 5, cols: 5 };

// 5×5 bands: low < 5, medium < 10, high < 15, else critical.
const MEDIUM_RATIO = 5 / 25;
const HIGH_RATIO = 10 / 25;
const CRITICAL_RATIO = 15 / 25;

export function defaultScoreLevels(rows: number, cols: number): RiskScoreLevelMapping[] {
    const max = rows * cols;
    const threshold = (numerator: number) => Math.ceil((max * numerator) / 25);

    return [
        { level: "LOW", minScore: 1 },
        { level: "MEDIUM", minScore: threshold(5) },
        { level: "HIGH", minScore: threshold(10) },
        { level: "CRITICAL", minScore: threshold(15) },
    ];
}

export function getRiskScoreLevel(
    score: number,
    matrixSize: RiskMatrixSize = DEFAULT_MATRIX_SIZE,
    scoreLevels?: ReadonlyArray<RiskScoreLevelMapping> | null,
): 0 | 1 | 2 | 3 {
    if (scoreLevels != null && scoreLevels.length > 0) {
        let chosen = scoreLevels[0].level;
        let bestMin = -1;
        for (const entry of scoreLevels) {
            if (entry.minScore <= score && entry.minScore >= bestMin) {
                bestMin = entry.minScore;
                chosen = entry.level;
            }
        }

        return LEVEL_INDEX[chosen];
    }

    const max = matrixSize.rows * matrixSize.cols;
    if (score >= max * CRITICAL_RATIO) {
        return 3;
    }
    if (score >= max * HIGH_RATIO) {
        return 2;
    }
    if (score >= max * MEDIUM_RATIO) {
        return 1;
    }
    return 0;
}

function getRiskSeverities(t: Translator, matrixSize: RiskMatrixSize) {
    const max = matrixSize.rows * matrixSize.cols;
    return [
        {
            min: max * CRITICAL_RATIO,
            variant: "danger",
            label: t("helpers.riskSeverity.critical"),
            bg: "bg-danger",
            color: "text-txt-danger",
        },
        {
            min: max * HIGH_RATIO,
            variant: "warning",
            label: t("helpers.riskSeverity.high"),
            bg: "bg-warning",
            color: "text-txt-warning",
        },
        {
            min: max * MEDIUM_RATIO,
            variant: "medium",
            label: t("helpers.riskSeverity.medium"),
            bg: "bg-medium",
            color: "text-txt-medium",
        },
        {
            min: 0,
            variant: "success",
            label: t("helpers.riskSeverity.low"),
            bg: "bg-success",
            color: "text-txt-success",
        },
    ] as const;
}

export function getSeverity(
    t: Translator,
    score: number,
    matrixSize: RiskMatrixSize = DEFAULT_MATRIX_SIZE,
    scoreLevels?: ReadonlyArray<RiskScoreLevelMapping> | null,
) {
    if (scoreLevels != null && scoreLevels.length > 0) {
        return getRiskSeverities(t, matrixSize)[
            3 - getRiskScoreLevel(score, matrixSize, scoreLevels)
        ];
    }

    return getRiskSeverities(t, matrixSize).find((s) => score >= s.min);
}
