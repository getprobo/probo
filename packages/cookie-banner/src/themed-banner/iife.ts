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

import { getConsent } from "../consent";
import { registerCookieBanner } from "./index";

registerCookieBanner();

const w = window as unknown as Record<string, unknown>;
if (!w.Probo) {
  w.Probo = {};
}
(w.Probo as Record<string, unknown>).consent = getConsent();

const script = document.currentScript as HTMLScriptElement | null;

if (script) {
  const bannerId = script.getAttribute("data-banner-id");
  const baseUrl = script.getAttribute("data-base-url");

  if (bannerId && baseUrl) {
    const mount = (): void => {
      const el = document.createElement("probo-cookie-banner");
      el.setAttribute("banner-id", bannerId);
      el.setAttribute("base-url", baseUrl);

      const position = script.getAttribute("data-position");
      if (position) {
        el.setAttribute("position", position);
      }

      const reopenWidget = script.getAttribute("data-reopen-widget");
      if (reopenWidget) {
        el.setAttribute("reopen-widget", reopenWidget);
      }

      const lang = script.getAttribute("data-lang");
      if (lang) {
        el.setAttribute("lang", lang.split("-")[0].toLowerCase());
      }

      document.body.appendChild(el);
    };

    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", mount);
    } else {
      mount();
    }
  }
}
