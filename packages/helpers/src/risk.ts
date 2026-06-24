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

export function getRiskImpacts(__: Translator) {
    return [
        {
            value: 1,
            label: __("Negligible"),
        },
        {
            value: 2,
            label: __("Low"),
        },
        {
            value: 3,
            label: __("Moderate"),
        },
        {
            value: 4,
            label: __("Significant"),
        },
        {
            value: 5,
            label: __("Catastrophic"),
        },
    ];
}

export function getTreatment(__: Translator, treatment?: string): string {
    switch (treatment) {
        case "MITIGATED":
            return __("Mitigate");
        case "ACCEPTED":
            return __("Accept");
        case "TRANSFERRED":
            return __("Transfer");
        case "AVOIDED":
            return __("Avoid");
        default:
            return __("Unknown");
    }
}

export function getRiskLikelihoods(__: Translator) {
    return [
        {
            value: 1,
            label: __("Improbable"),
        },
        {
            value: 2,
            label: __("Remote"),
        },
        {
            value: 3,
            label: __("Occasional"),
        },
        {
            value: 4,
            label: __("Probable"),
        },
        {
            value: 5,
            label: __("Frequent"),
        },
    ];
}

function getRiskSeverities(__: Translator) {
    return [
        {
            min: 15,
            variant: "danger",
            label: __("Critical"),
            bg: "bg-danger",
            color: "text-txt-danger",
        },
        {
            min: 5,
            variant: "warning",
            label: __("High"),
            bg: "bg-warning",
            color: "text-txt-warning",
        },
        {
            min: 0,
            variant: "neutral",
            label: __("Low"),
            bg: "bg-txt-quaternary",
            color: "text-txt-primary",
        },
    ] as const;
}

export function getSeverity(__: Translator, score: number) {
    return getRiskSeverities(__).find((s) => score >= s.min);
}
