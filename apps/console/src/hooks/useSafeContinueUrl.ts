import { useMemo } from "react";
import { useSearchParams } from "react-router";

export function useSafeContinueUrl(fallback?: string): URL {
  const [searchParams] = useSearchParams();

  const continueUrlParam = searchParams.get("continue");
  const safeContinueUrl = useMemo(() => {
    if (continueUrlParam) {
      let continueUrl: URL;
      try {
        continueUrl = new URL(continueUrlParam, window.location.origin);
      } catch {
        continueUrl = new URL(fallback ?? window.location.origin, window.location.origin);
      }
      return new URL(continueUrl.pathname + continueUrl.search, window.location.origin);
    } else {
      return new URL(fallback ?? window.location.origin, window.location.origin);
    }
  }, [continueUrlParam, fallback]);

  return safeContinueUrl;
}
