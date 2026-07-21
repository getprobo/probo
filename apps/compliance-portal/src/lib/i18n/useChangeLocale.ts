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

import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router";

import {
  replaceLocaleInPathname,
  type UrlLocale,
} from "./locale";
import { useLocale } from "./useLocale";
import { useUpdateLocale } from "./useUpdateLocale";

interface ChangeLocaleOptions {
  // When true, also persist the locale on the signed-in identity.
  persist?: boolean;
}

// Navigate to the same path under a new locale; optionally persist for signed-in users.
export function useChangeLocale() {
  const currentLocale = useLocale();
  const { pathname, search } = useLocation();
  const navigate = useNavigate();
  const [updateLocale, isUpdating] = useUpdateLocale();

  const changeLocale = useCallback(async (
    locale: UrlLocale,
    options: ChangeLocaleOptions = {},
  ) => {
    // Persist and navigate without sequencing them: awaiting the mutation
    // before navigation left a frame where Identity.locale already matched
    // the new choice but the URL still had the old prefix, flashing the
    // mismatch callout. updateLocale writes the store optimistically, and
    // flushSync applies the URL change in the same paint.
    if (options.persist) {
      void updateLocale(locale);
    }
    if (locale !== currentLocale) {
      void navigate(replaceLocaleInPathname(pathname, locale) + search, {
        flushSync: true,
      });
    }
  }, [currentLocale, navigate, pathname, search, updateLocale]);

  return [changeLocale, isUpdating] as const;
}
