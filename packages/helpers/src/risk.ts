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
