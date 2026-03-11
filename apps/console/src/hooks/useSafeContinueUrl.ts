import { useSearchParams } from "react-router";

export function useSafeContinueUrl(fallbackUrl?: URL): URL {
  fallbackUrl = new URL(fallbackUrl ?? window.location.origin, window.location.origin);
  const [searchParams] = useSearchParams();

  const continueUrlParam = searchParams.get("continue");
  let safeContinueUrl: URL;
  if (continueUrlParam) {
    let continueUrl: URL;
    try {
      continueUrl = new URL(continueUrlParam, window.location.origin);
    } catch {
      continueUrl = fallbackUrl;
    }
    safeContinueUrl = new URL(continueUrl.pathname + continueUrl.search, window.location.origin);
  } else {
    safeContinueUrl = fallbackUrl;
  }

  return safeContinueUrl;
}
