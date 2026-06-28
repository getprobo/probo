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

import { createInstance } from "i18next";
import { initReactI18next } from "react-i18next";

import { DEFAULT_NAMESPACE, globBackend } from "./backend";
import { resolveLanguage, SUPPORTED_LANGUAGES } from "./resolveLanguage";

// Build a dedicated instance rather than mutating i18next's global singleton.
// Initializing it through initReactI18next still registers it as the instance
// react-i18next's hooks read from, so no I18nextProvider is required.
const i18n = createInstance();

void i18n
  .use(globBackend)
  .use(initReactI18next)
  .init({
    lng: resolveLanguage(),
    fallbackLng: "en-US",
    supportedLngs: SUPPORTED_LANGUAGES,
    load: "currentOnly",
    defaultNS: DEFAULT_NAMESPACE,
    fallbackNS: DEFAULT_NAMESPACE,
    ns: [DEFAULT_NAMESPACE],
    interpolation: { escapeValue: false },
    react: { useSuspense: true },
  });

export { i18n };
