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
