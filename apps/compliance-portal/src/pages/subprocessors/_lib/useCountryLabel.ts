// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { useTranslation } from "react-i18next";

// Resolves a trust-center CountryCode to a localized display name. ISO 3166
// alpha-2 codes go through Intl.DisplayNames (locale-aware); the schema's two
// pseudo-regions (GLOBAL, EU) fall back to translated labels.
export function useCountryLabel(): (code: string) => string {
  const { t, i18n } = useTranslation("subprocessors");

  return useMemo(() => {
    const display = new Intl.DisplayNames([i18n.language], { type: "region" });

    return (code: string): string => {
      if (code === "GLOBAL") {
        return t("regions.global");
      }
      if (code === "EU") {
        return t("regions.eu");
      }
      try {
        return display.of(code) ?? code;
      } catch {
        return code;
      }
    };
  }, [i18n.language, t]);
}
