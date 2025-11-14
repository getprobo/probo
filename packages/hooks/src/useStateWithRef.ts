import { useCallback, useRef, useState } from "react";

/**
 * A useState hook that also returns a ref to the current state (usable in callbacks)
 */
export function useStateWithRef<T>(initialValue: T) {
    const [state, setState] = useState(initialValue);
    const ref = useRef(state);

    return [
        state,
        useCallback((v: T) => {
            setState(v);
            ref.current = v;
        }, []),
        ref,
    ] as const;
}
