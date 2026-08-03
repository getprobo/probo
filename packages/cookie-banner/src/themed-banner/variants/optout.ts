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
import { floatingCard } from "./shared";

// Opt-out (CCPA-style): trackers fire immediately; the visitor acknowledges or
// opts out. No preference panel. The statutory "Your Privacy Choices" reopen
// link is handled by <probo-settings-link> via layout.settings_link.
export function renderOptOut(position: string): string {
  return `
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
}
