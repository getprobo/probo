import { useEffect, useRef } from "react";

/**
 * Hook to handle cleanup after a delay
 *
 * Used for disposing the graphQL query when the component unmounts
 */
export function useCleanup(callback: () => void, delay: number) {
  const timer = useRef<ReturnType<typeof setTimeout>>(null);

  useEffect(() => {
    if (timer.current) {
      clearTimeout(timer.current);
    }
    return () => {
      timer.current = setTimeout(callback, delay);
    };
  }, [callback, delay]);
}
