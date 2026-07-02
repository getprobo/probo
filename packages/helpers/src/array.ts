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
