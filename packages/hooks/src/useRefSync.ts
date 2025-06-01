import { useRef } from "react";

export function useRefSync<T>(value: T) {
    const ref = useRef(value);
    ref.current = value;
    return ref;
}
