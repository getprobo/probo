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

// Injected into the themed-banner shadow so TCF can widen the preference
// panel without changing the GDPR category layout.
export const TCF_PANEL_STYLES = `<style>
  probo-banner .tcf-disclosures {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin: 0 0 16px;
  }

  probo-banner .tcf-disclosures .description {
    margin: 0;
  }

  probo-preference-panel.tcf-panel .card {
    max-width: 720px;
  }

  .tcf-section {
    font-size: calc(var(--_font-size) - 1px);
    font-weight: 500;
    color: var(--_text-secondary);
    padding: 16px 24px 8px;
  }

  .tcf-group {
    border-bottom: 1px solid var(--_border);
  }

  .tcf-subsection {
    font-size: var(--_font-size);
    font-weight: 600;
    padding: 12px 24px 4px;
  }

  .tcf-section-desc {
    color: var(--_text-secondary);
    margin: 0;
    padding: 0 24px 8px;
  }

  .tcf-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    padding: 8px 24px;
  }

  .tcf-row-id {
    flex-shrink: 0;
    min-width: 1.5em;
    color: var(--_text-secondary);
    font-variant-numeric: tabular-nums;
    font-weight: 500;
    margin-top: 1px;
  }

  .tcf-row .category-info {
    flex: 1;
    min-width: 0;
  }

  .tcf-vendor-details {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: 4px;
  }

  .tcf-vendor-line,
  .tcf-purpose-vendors {
    margin: 0;
    font-size: calc(var(--_font-size) - 1px);
    color: var(--_text-secondary);
  }

  .tcf-purpose-vendors {
    margin-top: 4px;
  }

  .tcf-control {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
  }

  .tcf-control-label {
    font-size: calc(var(--_font-size) - 2px);
    color: var(--_text-secondary);
    text-align: center;
    line-height: 1.2;
    max-width: 7.5em;
  }

  .panel-body > .tcf-group:last-child {
    border-bottom: none;
  }

  .tcf-storage {
    color: var(--_text-secondary);
    margin: 0;
    padding: 0 24px 16px;
  }
</style>`;
