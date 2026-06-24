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
