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

export const measureStates = [
    "IMPLEMENTED",
    "IN_PROGRESS",
    "NOT_APPLICABLE",
    "NOT_STARTED",
    "UNKNOWN",
    "NOT_IMPLEMENTED",
] as const;

export function getMeasureStateLabel(__: Translator, state: string) {
    switch (state) {
        case "IMPLEMENTED":
            return __("Implemented");
        case "IN_PROGRESS":
            return __("In Progress");
        case "NOT_APPLICABLE":
            return __("Not Applicable");
        case "NOT_STARTED":
            return __("Not Started");
        case "UNKNOWN":
            return __("Unknown");
        case "NOT_IMPLEMENTED":
            return __("Not Implemented");
        default:
            return __("Unknown");
    }
}
