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
