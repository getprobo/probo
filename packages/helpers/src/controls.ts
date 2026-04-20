// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

export const controlMaturityLevels = [
    "NONE",
    "INITIAL",
    "MANAGED",
    "DEFINED",
    "QUANTITATIVELY_MANAGED",
    "OPTIMIZING",
] as const;

export type ControlMaturityLevel = (typeof controlMaturityLevels)[number];

export function getControlMaturityLevelLabel(__: Translator, level: string) {
    switch (level) {
        case "NONE":
            return __("0 - None");
        case "INITIAL":
            return __("1 - Initial");
        case "MANAGED":
            return __("2 - Managed");
        case "DEFINED":
            return __("3 - Defined");
        case "QUANTITATIVELY_MANAGED":
            return __("4 - Quantitatively Managed");
        case "OPTIMIZING":
            return __("5 - Optimizing");
    }
}
