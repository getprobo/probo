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

import { type LoaderFunctionArgs, redirect } from "react-router";

import {
  isUrlLocale,
  localizedPath,
  resolveUrlLocale,
} from "./locale";

// Ensures `/:lang` is a supported short tag. Legacy unprefixed URLs
// (/documents) and unknown two-letter tags are rewritten to a guessed locale.
export function localeLayoutLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const appPath = url.pathname || "/";
  const segments = appPath.split("/").filter(Boolean);
  const first = segments[0];

  if (isUrlLocale(first)) {
    return null;
  }

  const guess = resolveUrlLocale();

  if (segments.length === 0) {
    // eslint-disable-next-line @typescript-eslint/only-throw-error -- react-router redirect
    throw redirect(`/${guess}${url.search}`);
  }

  // Unknown two-letter tag (e.g. /xx/documents) → drop it and keep the rest.
  if (/^[a-z]{2}$/.test(first)) {
    const rest = segments.slice(1);
    const path = rest.length === 0 ? "/" : `/${rest.join("/")}`;
    // eslint-disable-next-line @typescript-eslint/only-throw-error -- react-router redirect
    throw redirect(localizedPath(guess, path) + url.search);
  }

  // First segment is a real route (documents, updates, …) — prefix locale.
  // eslint-disable-next-line @typescript-eslint/only-throw-error -- react-router redirect
  throw redirect(localizedPath(guess, appPath) + url.search);
}
