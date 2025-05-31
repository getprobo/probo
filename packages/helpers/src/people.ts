type Translator = (s: string) => string;

export const peopleRoles = [
    "EMPLOYEE",
    "CONTRACTOR",
    "SERVICE_ACCOUNT",
] as const;

export function getRoles(__: Translator) {
    return [
        {
            value: "EMPLOYEE",
            label: __("Employee"),
        },
        {
            value: "CONTRACTOR",
            label: __("Contractor"),
        },
        {
            value: "SERVICE_ACCOUNT",
            label: __("Service account"),
        },
    ];
}

export function getRole(__: Translator, role?: string): string {
    switch (role) {
        case "EMPLOYEE":
            return __("Employee");
        case "CONTRACTOR":
            return __("Contractor");
        case "SERVICE_ACCOUNT":
            return __("Service account");
        default:
            return __("Unknown");
    }
}
