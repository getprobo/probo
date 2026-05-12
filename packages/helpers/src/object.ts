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

/**
 * A type safe version of Object.keys
 */
export function objectKeys<T extends Record<string, unknown>>(object: T) {
    return Object.keys(object) as (keyof T)[];
}

export function objectEntries<T extends Record<string, unknown>>(object: T) {
    return Object.entries(object) as [keyof T, T[keyof T]][];
}

/**
 * Trims string values and converts empty strings to null in form data objects
 */
export function cleanFormData<T extends Record<string, any>>(data: T): T {
    return Object.fromEntries(
        Object.entries(data).map(([k, v]) => {
            const trimmed = typeof v === 'string' ? v.trim() : v;
            return [k, trimmed === "" ? null : trimmed];
        })
    ) as T;
}
