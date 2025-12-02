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
