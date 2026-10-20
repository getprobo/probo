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

export const trackerTypeBadges = {
  COOKIE: { color: "amber" as const, labelKey: "cookie", shortKey: "cookieShort", variant: "soft" as const },
  LOCAL_STORAGE: { color: "sky" as const, labelKey: "localStorage", shortKey: "localStorageShort", variant: "soft" as const },
  SESSION_STORAGE: { color: "indigo" as const, labelKey: "sessionStorage", shortKey: "sessionStorageShort", variant: "soft" as const },
  INDEXED_DB: { color: "green" as const, labelKey: "indexedDb", shortKey: "indexedDbShort", variant: "soft" as const },
  CACHE_STORAGE: { color: "neutral" as const, labelKey: "cacheStorage", shortKey: "cacheStorageShort", variant: "outline" as const },
};

export const cookieSourceBadges = {
  SCRIPT: { color: "sky" as const, labelKey: "script", variant: "soft" as const },
  PRE_EXISTING: { color: "neutral" as const, labelKey: "preExisting", variant: "outline" as const },
  HTTP: { color: "neutral" as const, labelKey: "http", variant: "soft" as const },
  EXTENSION: { color: "amber" as const, labelKey: "extension", variant: "soft" as const },
};
