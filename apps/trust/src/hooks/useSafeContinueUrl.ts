// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
