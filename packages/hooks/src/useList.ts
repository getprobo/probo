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

import { useCallback, useState } from "react";

export function useList<T>(items: T[]) {
    const [list, setList] = useState(items);
    const push = useCallback(
        (item: T) => setList((prev) => [...prev, item]),
        [],
    );
    const remove = useCallback(
        (item: T) => setList((prev) => prev.filter((i) => i !== item)),
        [],
    );
    const toggle = useCallback(
        (item: T) =>
            setList((prev) =>
                prev.includes(item)
                    ? prev.filter((i) => i !== item)
                    : [...prev, item],
            ),
        [],
    );
    const reset = useCallback((items: T[]) => setList(items), []);
    const clear = useCallback(() => setList([]), []);
    return { list, push, remove, toggle, reset, clear };
}
