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

import { useCallback, useRef, useState } from "react";

/**
 * A useState hook that also returns a ref to the current state (usable in callbacks)
 */
export function useStateWithRef<T>(initialValue: T) {
    const [state, setState] = useState<T>(initialValue);
    const ref = useRef(state);

    return [
        state,
        useCallback((v: T | ((prevState: T) => T)) => {
            setState(prev => {
                const nextState = typeof v === "function"
                    ? (v as (prevState: T) => T)(prev) : v;
                ref.current = nextState;
                return nextState;
            });
        }, []),
        ref,
    ] as const;
}
