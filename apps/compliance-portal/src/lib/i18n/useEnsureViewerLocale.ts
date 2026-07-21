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

import { useEffect, useRef } from "react";
import { graphql, useFragment } from "react-relay";

import type { useEnsureViewerLocale_identity$key } from "./__generated__/useEnsureViewerLocale_identity.graphql";
import { useLocale } from "./useLocale";
import { useUpdateLocale } from "./useUpdateLocale";

const ensureViewerLocaleFragment = graphql`
  fragment useEnsureViewerLocale_identity on Identity {
    locale
  }
`;

// Seeds Identity.locale once when it is still null, using the current URL
// locale. Never overwrites an existing preference (shared-link safe).
export function useEnsureViewerLocale(
  identityKey: useEnsureViewerLocale_identity$key | null,
) {
  const identity = useFragment(ensureViewerLocaleFragment, identityKey);
  const urlLocale = useLocale();
  const [updateLocale] = useUpdateLocale();
  const firedRef = useRef(false);

  useEffect(() => {
    if (identity == null || identity.locale != null || firedRef.current) {
      return;
    }

    firedRef.current = true;
    void updateLocale(urlLocale, { errorToast: false }).catch(() => {
      // Allow a later mount / navigation to retry if the seed failed.
      firedRef.current = false;
    });
  }, [identity, updateLocale, urlLocale]);
}
