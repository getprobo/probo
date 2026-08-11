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

import { BRANDING } from "../../html";
import { getDismissLabel } from "../../i18n";
import type { BannerConfig } from "../../types";
import { esc, floatingCard } from "./shared";

// Notice-only (implied consent): trackers fire immediately; the visitor merely
// acknowledges an informational banner. No reject, customize, or panel.
//
// The dismiss button records ACKNOWLEDGE, so it must never borrow the
// "Accept all" wording. When a translation omits `button_dismiss` it keeps a
// localized "Got it" default (applyTexts leaves the inlined text untouched).
export function renderNotice(config: BannerConfig, position: string): string {
  const dismissDefault = esc(getDismissLabel(config.language));

  return `
    <probo-banner>
      ${floatingCard(
        position,
        { labelledby: "probo-banner-title", describedby: "probo-banner-desc" },
        `
        <p class="title" id="probo-banner-title" data-text="banner_title_notice" data-text-fallback="banner_title"></p>
        <p class="description" id="probo-banner-desc" data-text="banner_description_notice" data-text-fallback="banner_description"></p>
        <div class="buttons">
          <probo-acknowledge-button><button class="btn btn-primary" data-text="button_dismiss">${dismissDefault}</button></probo-acknowledge-button>
        </div>
        ${BRANDING}`,
      )}
    </probo-banner>`;
}
