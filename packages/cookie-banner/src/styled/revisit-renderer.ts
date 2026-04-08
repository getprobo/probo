// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import type { ThemeConfig, WidgetStrings } from "../headless/types";
import { defaultTheme } from "../headless/api";

export function renderRevisitIcon(
  root: ShadowRoot,
  onClick: () => void,
  strings: WidgetStrings,
  theme?: ThemeConfig,
): void {
  const t = theme ?? defaultTheme;

  const positionCSS = t.revisit_position === "bottom-right"
    ? "right: 20px;"
    : "left: 20px;";

  const style = document.createElement("style");
  style.textContent = `
    .probo-revisit {
      position: fixed; bottom: 20px; ${positionCSS} z-index: 2147483646;
      width: 44px; height: 44px; border-radius: 50%;
      background: ${t.primary_color}; color: ${t.primary_text_color}; border: none;
      cursor: pointer; display: flex; align-items: center; justify-content: center;
      box-shadow: 0 2px 8px rgba(0,0,0,0.15); transition: transform 0.2s;
      font-family: ${t.font_family};
    }
    .probo-revisit:hover { transform: scale(1.1); }
    .probo-revisit svg { width: 20px; height: 20px; }
  `;
  root.appendChild(style);

  const btn = document.createElement("button");
  btn.className = "probo-revisit";
  btn.title = strings.cookiePreferencesTooltip;
  btn.innerHTML = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="8" cy="10" r="1" fill="currentColor"/><circle cx="15" cy="8" r="1" fill="currentColor"/><circle cx="10" cy="15" r="1" fill="currentColor"/><circle cx="16" cy="13" r="1" fill="currentColor"/></svg>`;
  btn.addEventListener("click", onClick);
  root.appendChild(btn);
}
