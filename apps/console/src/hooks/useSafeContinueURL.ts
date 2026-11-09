// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { useMemo } from "react";
import { useSearchParams } from "react-router";

// "//evil.example" stays on this origin in the first parse, then becomes scheme-relative.
export function safeContinueURL(
  continueURLParam: string | null,
  origin: string,
  fallback?: string,
): URL {
  if (!continueURLParam) {
    return sameOriginURL(fallback, origin);
  }

  let continueURL: URL;
  try {
    continueURL = new URL(continueURLParam, origin);
  } catch {
    return sameOriginURL(fallback, origin);
  }

  let candidate: URL;
  try {
    candidate = new URL(continueURL.pathname + continueURL.search, origin);
  } catch {
    return sameOriginURL(fallback, origin);
  }

  if (candidate.origin !== origin) {
    return sameOriginURL(fallback, origin);
  }

  return candidate;
}

function sameOriginURL(value: string | undefined, origin: string): URL {
  try {
    const fallbackURL = new URL(value ?? origin, origin);
    if (fallbackURL.origin === origin) {
      return fallbackURL;
    }
  } catch {
    // Invalid fallback.
  }

  return new URL(origin);
}

export function useSafeContinueURL(fallback?: string): URL {
  const [searchParams] = useSearchParams();
  const continueURLParam = searchParams.get("continue");

  return useMemo(
    () => safeContinueURL(continueURLParam, window.location.origin, fallback),
    [continueURLParam, fallback],
  );
}
