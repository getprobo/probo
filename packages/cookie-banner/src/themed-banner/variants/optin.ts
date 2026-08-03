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
import { CATEGORY_LIST, floatingCard } from "./shared";

// Opt-in (GDPR-style): non-necessary trackers stay blocked until the visitor
// consents. Offers accept / reject / customize and a full preference panel.
export function renderOptIn(position: string): string {
  const banner = `
    <probo-banner>
      ${floatingCard(
        position,
        { labelledby: "probo-banner-title", describedby: "probo-banner-desc" },
        `
        <p class="title" id="probo-banner-title" data-text="banner_title"></p>
        <p class="description" id="probo-banner-desc" data-text="banner_description"></p>
        <div class="buttons">
          <probo-accept-button><button class="btn btn-primary" data-text="button_accept_all"></button></probo-accept-button>
          <probo-reject-button><button class="btn" data-text="button_reject_all"></button></probo-reject-button>
          <probo-customize-button><button class="btn btn-link" data-text="button_customize"></button></probo-customize-button>
        </div>
        ${BRANDING}`,
      )}
    </probo-banner>`;

  const panel = `
    <probo-preference-panel>
      ${floatingCard(
        position,
        { labelledby: "probo-panel-title", describedby: "probo-panel-desc" },
        `
        <div class="panel-header">
          <div class="panel-header-title">
            <p class="title" id="probo-panel-title" style="margin:0" data-text="panel_title"></p>
            <button class="panel-close" data-action="back" data-aria-text="aria_close">
              ${CLOSE_ICON}
            </button>
          </div>
          <p class="description" id="probo-panel-desc" data-text="panel_description"></p>
        </div>
        ${CATEGORY_LIST}
        <div class="footer">
          <div class="buttons">
            <probo-accept-button><button class="btn btn-primary" data-text="button_accept_all"></button></probo-accept-button>
            <probo-reject-button><button class="btn" data-text="button_reject_all"></button></probo-reject-button>
            <probo-save-button>
              <button class="btn btn-link" style="flex:1" data-text="button_save"></button>
            </probo-save-button>
          </div>
          ${BRANDING}
        </div>`,
      )}
    </probo-preference-panel>`;

  return banner + panel;
}
