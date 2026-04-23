// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { registerThemedBanner } from "./index";

registerThemedBanner();

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
