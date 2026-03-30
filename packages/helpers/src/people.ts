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
