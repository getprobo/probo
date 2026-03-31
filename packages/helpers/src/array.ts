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

export function times<T>(n: number, cb: (i: number) => T): T[] {
    return Array.from({ length: n }, (_, i) => cb(i));
}

export function groupBy<T>(
    arr: T[],
    key: (item: T) => string,
): Record<string, T[]> {
    return arr.reduce(
        (acc, item) => {
            const k = key(item);
            if (!acc[k]) {
                acc[k] = [];
            }
            acc[k].push(item);
            return acc;
        },
        {} as Record<string, T[]>,
    );
}

/**
 * Check that a value is empty (null, undefined, empty string, empty array, empty object)
 */
export function isEmpty(v: unknown): boolean {
    if (Array.isArray(v)) {
        return v.find((v) => !isEmpty(v)) === undefined;
    }
    return !v;
}
