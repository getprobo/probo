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

function escapeAttr(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

export function headlessRootHTML(
  bannerId: string,
  baseUrl: string,
  gcmEnabled: boolean,
): string {
  return `
      <style>probo-banner, probo-preference-panel, probo-privacy-choices { display: block !important; }</style>
      <probo-cookie-banner-root banner-id="${escapeAttr(bannerId)}" base-url="${escapeAttr(baseUrl)}" gcm-enabled="${gcmEnabled ? "true" : "false"}">
        <probo-banner>
          <div style="border:2px solid #333;padding:12px;margin-bottom:8px;">
            <strong>[probo-banner]</strong>
            <div style="margin-top:8px;">
              <probo-acknowledge-button><button>Acknowledge</button></probo-acknowledge-button>
              <probo-accept-button><button style="margin-left:8px;">Accept All</button></probo-accept-button>
              <probo-reject-button><button style="margin-left:8px;">Reject All</button></probo-reject-button>
              <probo-customize-button><button style="margin-left:8px;">Customize</button></probo-customize-button>
            </div>
          </div>
        </probo-banner>

        <probo-preference-panel>
          <div style="border:2px dashed #666;padding:12px;margin-bottom:8px;">
            <strong>[probo-preference-panel]</strong>
            <probo-category-list>
              <template>
                <div style="border:1px solid #aaa;padding:8px;margin:4px 0;">
                  <span data-slot="name" style="font-weight:bold;"></span>:
                  <span data-slot="description"></span>
                  <probo-category-toggle>
                    <label style="margin-left:8px;"><input type="checkbox" /> toggle</label>
                  </probo-category-toggle>
                  <probo-cookie-list hidden>
                    <template>
                      <div style="padding:4px 0 4px 16px;font-size:13px;">
                        <span data-slot="name" style="font-weight:bold;"></span>
                        &mdash; <span data-slot="description"></span>
                      </div>
                    </template>
                  </probo-cookie-list>
                </div>
              </template>
            </probo-category-list>
            <div style="margin-top:8px;">
              <probo-accept-button><button>Accept All</button></probo-accept-button>
              <probo-reject-button><button style="margin-left:8px;">Reject All</button></probo-reject-button>
              <probo-save-button><button style="margin-left:8px;">Save Preferences</button></probo-save-button>
            </div>
          </div>
        </probo-preference-panel>

        <probo-privacy-choices>
          <div style="border:2px solid #1d4ed8;padding:12px;margin-bottom:8px;">
            <strong>[probo-privacy-choices]</strong>
            <p style="margin:8px 0;font-size:14px;">
              Right to opt out of sale/sharing and right to limit sensitive
              personal information (CCPA).
            </p>
            <probo-reject-button>
              <button>Do Not Sell or Share My Personal Information</button>
            </probo-reject-button>
          </div>
        </probo-privacy-choices>
      </probo-cookie-banner-root>
    `;
}
