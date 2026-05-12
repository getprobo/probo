// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
