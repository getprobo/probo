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

export const THEMED_STYLES = `
  :host {
    --_font: var(--probo-font-family, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif);
    --_bg: var(--probo-bg, #ffffff);
    --_text: var(--probo-text, #1a1a1a);
    --_text-secondary: var(--probo-text-secondary, #555555);
    --_border: var(--probo-border, #e0e0e0);
    --_radius: var(--probo-radius, 12px);
    --_shadow: var(--probo-shadow, 0 4px 24px rgba(0, 0, 0, 0.12));
    --_accent: var(--probo-accent, #1a1a1a);
    --_accent-text: var(--probo-accent-text, #ffffff);
    --_overlay: var(--probo-overlay, rgba(0, 0, 0, 0.4));
    --_z-index: var(--probo-z-index, 2147483646);

    all: initial;
    font-family: var(--_font);
    color: var(--_text);
    font-size: 14px;
    line-height: 1.5;
    box-sizing: border-box;
  }

  *, *::before, *::after {
    box-sizing: border-box;
  }

  .overlay {
    position: fixed;
    inset: 0;
    background: var(--_overlay);
    z-index: var(--_z-index);
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding: 16px;
  }

  .card {
    background: var(--_bg);
    border-radius: var(--_radius);
    box-shadow: var(--_shadow);
    max-width: 520px;
    width: 100%;
    padding: 24px;
    max-height: 85vh;
    overflow-y: auto;
  }

  .title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 8px;
  }

  .description {
    color: var(--_text-secondary);
    margin: 0 0 20px;
    font-size: 14px;
  }

  .description a {
    color: var(--_accent);
    text-decoration: underline;
  }

  .buttons {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .btn {
    flex: 1;
    min-width: 0;
    padding: 10px 16px;
    border-radius: 8px;
    border: 1px solid var(--_border);
    background: var(--_bg);
    color: var(--_text);
    font-family: var(--_font);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.15s, border-color 0.15s;
    white-space: nowrap;
  }

  .btn:hover {
    border-color: var(--_accent);
  }

  .btn-primary {
    background: var(--_accent);
    color: var(--_accent-text);
    border-color: var(--_accent);
  }

  .btn-primary:hover {
    opacity: 0.9;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .panel-back {
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    color: var(--_text);
    font-family: var(--_font);
    font-size: 14px;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .panel-back:hover {
    color: var(--_accent);
  }

  probo-category-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 20px;
  }

  probo-category {
    display: block;
    border: 1px solid var(--_border);
    border-radius: 8px;
    padding: 12px 16px;
  }

  .category-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .category-info {
    flex: 1;
    min-width: 0;
  }

  .category-name {
    font-weight: 500;
    font-size: 14px;
  }

  .category-description {
    color: var(--_text-secondary);
    font-size: 13px;
    margin-top: 2px;
  }

  .category-required {
    font-size: 12px;
    color: var(--_text-secondary);
    font-style: italic;
  }

  .toggle {
    position: relative;
    width: 44px;
    height: 24px;
    flex-shrink: 0;
  }

  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
    position: absolute;
  }

  .toggle-track {
    position: absolute;
    inset: 0;
    background: var(--_border);
    border-radius: 12px;
    cursor: pointer;
    transition: background 0.2s;
  }

  .toggle-track::after {
    content: "";
    position: absolute;
    top: 2px;
    left: 2px;
    width: 20px;
    height: 20px;
    background: white;
    border-radius: 50%;
    transition: transform 0.2s;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }

  .toggle input:checked + .toggle-track {
    background: var(--_accent);
  }

  .toggle input:checked + .toggle-track::after {
    transform: translateX(20px);
  }

  .toggle input:disabled + .toggle-track {
    opacity: 0.5;
    cursor: not-allowed;
  }

  probo-cookie-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px solid var(--_border);
  }

  .cookie-item {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    font-size: 12px;
    gap: 8px;
  }

  .cookie-name {
    font-weight: 500;
    font-family: monospace;
    flex-shrink: 0;
  }

  .cookie-duration {
    color: var(--_text-secondary);
    flex-shrink: 0;
  }

  [hidden] {
    display: none !important;
  }

  @media (min-width: 640px) {
    .overlay {
      align-items: center;
    }
  }
`;
