// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

export function getTrackerTypeBadge(type: string, __: Translator): Badge {
  switch (type) {
    case "COOKIE": return { label: __("Cookie"), variant: "warning" };
    case "LOCAL_STORAGE": return { label: __("localStorage"), variant: "info" };
    case "SESSION_STORAGE": return { label: __("sessionStorage"), variant: "highlight" };
    case "INDEXED_DB": return { label: __("IndexedDB"), variant: "success" };
    case "CACHE_STORAGE": return { label: __("Cache Storage"), variant: "outline" };
    default: return { label: type, variant: "neutral" };
  }
}

export function getTrackerSourceBadge(source: string, __: Translator): Badge {
  switch (source) {
    case "SCRIPT": return { label: __("Script"), variant: "info" };
    case "PRE_EXISTING": return { label: __("Pre-existing"), variant: "outline" };
    case "HTTP": return { label: __("HTTP"), variant: "neutral" };
    case "EXTENSION": return { label: __("Extension"), variant: "warning" };
    default: return { label: source, variant: "neutral" };
  }
}
