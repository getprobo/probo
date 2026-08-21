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

const DEFAULT_MATRIX_SIZE: RiskMatrixSize = { rows: 5, cols: 5 };

// 5×5 bands: low < 5, high < 15, else critical. Scale by max product.
const HIGH_RATIO = 5 / 25;
const CRITICAL_RATIO = 15 / 25;

export function getRiskScoreLevel(
    score: number,
    matrixSize: RiskMatrixSize = DEFAULT_MATRIX_SIZE,
): 0 | 1 | 2 {
    const max = matrixSize.rows * matrixSize.cols;
    if (score >= max * CRITICAL_RATIO) {
        return 2;
    }
    if (score >= max * HIGH_RATIO) {
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
            min: 0,
            variant: "neutral",
            label: t("helpers.riskSeverity.low"),
            bg: "bg-txt-quaternary",
            color: "text-txt-primary",
        },
    ] as const;
}

export function getSeverity(
    t: Translator,
    score: number,
    matrixSize: RiskMatrixSize = DEFAULT_MATRIX_SIZE,
) {
    return getRiskSeverities(t, matrixSize).find((s) => score >= s.min);
}
