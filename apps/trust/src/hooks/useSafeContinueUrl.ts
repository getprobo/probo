import { useMemo } from "react";
import { useSearchParams } from "react-router";

import { getPathPrefix } from "#/utils/pathPrefix";

export function useSafeContinueUrl(): URL {
  const [searchParams] = useSearchParams();

  const continueUrlParam = searchParams.get("continue");
  const prefix = getPathPrefix();
  const fallback = window.location.origin + (prefix || "/");

  const safeContinueUrl = useMemo(() => {
    if (continueUrlParam) {
      let continueUrl: URL;
      try {
        continueUrl = new URL(continueUrlParam, window.location.origin);
      } catch {
        return new URL(fallback, window.location.origin);
      }
      if (
        continueUrl.origin === window.location.origin
        && continueUrl.pathname.startsWith(`${prefix}/`)
      ) {
        return new URL(
          continueUrl.pathname + continueUrl.search,
          window.location.origin,
        );
      }
      return new URL(fallback, window.location.origin);
    }
    return new URL(fallback, window.location.origin);
  }, [continueUrlParam, fallback, prefix]);

  return safeContinueUrl;
}
