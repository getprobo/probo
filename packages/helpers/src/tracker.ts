// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type Translator = (s: string) => string;

type BadgeVariant
  = | "warning"
    | "info"
    | "highlight"
    | "success"
    | "outline"
    | "neutral";

type Badge = {
  label: string;
  variant: BadgeVariant;
};

export function getTrackerTypeBadge(type: string, t: Translator): Badge {
  switch (type) {
    case "COOKIE": return { label: t("helpers.trackerType.cookie"), variant: "warning" };
    case "LOCAL_STORAGE": return { label: t("helpers.trackerType.localStorage"), variant: "info" };
    case "SESSION_STORAGE": return { label: t("helpers.trackerType.sessionStorage"), variant: "highlight" };
    case "INDEXED_DB": return { label: t("helpers.trackerType.indexedDb"), variant: "success" };
    case "CACHE_STORAGE": return { label: t("helpers.trackerType.cacheStorage"), variant: "outline" };
    default: return { label: type, variant: "neutral" };
  }
}

export function getTrackerSourceBadge(source: string, t: Translator): Badge {
  switch (source) {
    case "SCRIPT": return { label: t("helpers.trackerSource.script"), variant: "info" };
    case "PRE_EXISTING": return { label: t("helpers.trackerSource.preExisting"), variant: "outline" };
    case "HTTP": return { label: t("helpers.trackerSource.http"), variant: "neutral" };
    case "EXTENSION": return { label: t("helpers.trackerSource.extension"), variant: "warning" };
    default: return { label: source, variant: "neutral" };
  }
}
