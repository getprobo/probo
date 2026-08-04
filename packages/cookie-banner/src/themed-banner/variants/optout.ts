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

import { BRANDING, CLOSE_ICON } from "../../html";
import { floatingCard } from "./shared";

// Opt-out (CCPA-style): trackers fire immediately; the visitor acknowledges or
// opts out. Compact banner for first paint / non-CCPA reopen; Privacy Choices
// panel for CCPA's Alternative Opt-out Link (layout.reopen_state). The statutory
// "Your Privacy Choices" link is handled by <probo-settings-link>.
export function renderOptOut(position: string): string {
  const banner = `
    <probo-banner>
      ${floatingCard(
        position,
        { labelledby: "probo-banner-title", describedby: "probo-banner-desc" },
        `
        <p class="title" id="probo-banner-title" data-text="banner_title_opt_out" data-text-fallback="banner_title"></p>
        <p class="description" id="probo-banner-desc" data-text="banner_description_opt_out" data-text-fallback="banner_description"></p>
        <div class="buttons">
          <probo-accept-button><button class="btn btn-primary" data-text="button_acknowledge" data-text-fallback="button_accept_all"></button></probo-accept-button>
          <probo-reject-button><button class="btn" data-text="button_opt_out" data-text-fallback="button_reject_all"></button></probo-reject-button>
        </div>
        ${BRANDING}`,
      )}
    </probo-banner>`;

  const privacyChoices = `
    <probo-privacy-choices>
      ${floatingCard(
        position,
        {
          labelledby: "probo-privacy-choices-title",
          describedby: "probo-privacy-choices-desc",
        },
        `
        <div class="panel-header">
          <div class="panel-header-title">
            <p class="title" id="probo-privacy-choices-title" style="margin:0" data-text="privacy_choices_title"></p>
            <button class="panel-close" data-action="close-privacy-choices" data-aria-text="aria_close">
              ${CLOSE_ICON}
            </button>
          </div>
          <p class="description" id="probo-privacy-choices-desc" data-text="privacy_choices_intro"></p>
        </div>
        <div class="privacy-choices-body">
          <section class="privacy-choices-section">
            <p class="privacy-choices-section-title" data-text="privacy_choices_sale_title"></p>
            <p class="description" data-text="privacy_choices_sale_description"></p>
            <div class="buttons">
              <probo-reject-button><button class="btn" data-text="button_opt_out" data-text-fallback="button_reject_all"></button></probo-reject-button>
            </div>
          </section>
          <section class="privacy-choices-section privacy-choices-section-spi">
            <p class="privacy-choices-section-title" data-text="privacy_choices_spi_title"></p>
            <p class="description" data-text="privacy_choices_spi_description"></p>
          </section>
        </div>
        <div class="footer">
          ${BRANDING}
        </div>`,
      )}
    </probo-privacy-choices>`;

  return banner + privacyChoices;
}
