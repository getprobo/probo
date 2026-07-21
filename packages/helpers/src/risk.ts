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

export function getRiskImpacts(t: Translator) {
    return [
        {
            value: 1,
            label: t("helpers.riskImpact.negligible"),
        },
        {
            value: 2,
            label: t("helpers.riskImpact.low"),
        },
        {
            value: 3,
            label: t("helpers.riskImpact.moderate"),
        },
        {
            value: 4,
            label: t("helpers.riskImpact.significant"),
        },
        {
            value: 5,
            label: t("helpers.riskImpact.catastrophic"),
        },
    ];
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

export function getRiskLikelihoods(t: Translator) {
    return [
        {
            value: 1,
            label: t("helpers.riskLikelihood.improbable"),
        },
        {
            value: 2,
            label: t("helpers.riskLikelihood.remote"),
        },
        {
            value: 3,
            label: t("helpers.riskLikelihood.occasional"),
        },
        {
            value: 4,
            label: t("helpers.riskLikelihood.probable"),
        },
        {
            value: 5,
            label: t("helpers.riskLikelihood.frequent"),
        },
    ];
}

function getRiskSeverities(t: Translator) {
    return [
        {
            min: 15,
            variant: "danger",
            label: t("helpers.riskSeverity.critical"),
            bg: "bg-danger",
            color: "text-txt-danger",
        },
        {
            min: 5,
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

export function getSeverity(t: Translator, score: number) {
    return getRiskSeverities(t).find((s) => score >= s.min);
}
