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
    --_z-index: var(--probo-z-index, 2147483646);
    --_btn-radius: var(--probo-btn-radius, 8px);
    --_font-size: var(--probo-font-size, 14px);

    all: initial;
    font-family: var(--_font);
    color: var(--_text);
    font-size: var(--_font-size);
    line-height: 1.5;
    box-sizing: border-box;
  }

  *, *::before, *::after {
    box-sizing: border-box;
  }

  .floating {
    position: fixed;
    z-index: var(--_z-index);
    padding: 24px;
    max-width: 100vw;
    display: flex;
    pointer-events: none;
  }

  .floating[data-position="bottom-left"] {
    bottom: 0;
    left: 0;
  }

  .floating[data-position="bottom-right"] {
    bottom: 0;
    right: 0;
  }

  .floating[data-position="bottom-center"] {
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
  }

  .floating[data-position="top-left"] {
    top: 0;
    left: 0;
  }

  .floating[data-position="top-right"] {
    top: 0;
    right: 0;
  }

  .floating[data-position="top-center"] {
    top: 0;
    left: 50%;
    transform: translateX(-50%);
  }

  .card {
    background: var(--_bg);
    border-radius: var(--_radius);
    box-shadow: var(--_shadow);
    width: 100%;
    pointer-events: auto;
  }

  probo-banner .card {
    max-width: 450px;
    padding: 24px 24px 12px 24px;
  }

  probo-banner .buttons {
    padding-bottom: 12px;
  }

  probo-preference-panel .card {
    max-width: 520px;
    max-height: 75vh;
    display: flex;
    flex-direction: column;
  }

  probo-preference-panel probo-category-list {
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .title {
    font-size: calc(var(--_font-size) + 2px);
    font-weight: 600;
    margin: 0 0 8px;
  }

  .description {
    color: var(--_text-secondary);
    margin: 0 0 20px;
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
    padding: 8px 10px;
    border-radius: var(--_btn-radius);
    border: 1px solid var(--_border);
    background: color-mix(in srgb, var(--_text) 8%, var(--_bg));
    color: var(--_text);
    font-family: var(--_font);
    font-size: var(--_font-size);
    font-weight: 500;
    line-height: normal;
    cursor: pointer;
    transition: background 0.15s, border-color 0.15s;
    white-space: nowrap;
  }

  .btn:hover {
    opacity: 0.8;
  }

  .btn-link {
    background: transparent;
    border: none;
    color: var(--_accent);
    text-decoration: underline;
    padding: 8px 0;
  }

  .btn-primary {
    background: var(--_accent);
    color: var(--_accent-text);
    border-color: var(--_accent);
  }

  .panel-header {
    padding: 24px;
    border-bottom: 1px solid var(--_border);
  }

  .panel-header .description {
    margin: 0;
  }

  .panel-header-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 0 0 8px;
  }

  .panel-close {
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    color: var(--_text-secondary);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .panel-close:hover {
    color: var(--_text);
  }

  probo-category-list {
    display: flex;
    flex-direction: column;
  }

  probo-preference-panel .footer {
    border-top: 1px solid var(--_border);
    padding: 10px 24px;
  }

  probo-preference-panel .buttons {
    padding: 10px 0;
  }

  probo-category {
    display: block;
    border-bottom: 1px solid var(--_border);
    padding: 12px 40px;
    position: relative;
  }

  probo-category:last-child {
    border-bottom: none;
  }

  .category-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .category-info {
    flex: 1;
    min-width: 0;
  }

  .category-name {
    font-weight: 500;
  }

  .category-description {
    color: var(--_text-secondary);
    font-size: calc(var(--_font-size) - 1px);
    margin-top: 2px;
  }

  .toggle {
    position: relative;
    display: inline-block;
    width: 34px;
    height: 18px;
    flex-shrink: 0;
    margin-top: 2px;
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
    border-radius: 9px;
    cursor: pointer;
    transition: background 0.2s;
  }

  .toggle-track::after {
    content: "";
    position: absolute;
    top: 2px;
    left: 2px;
    width: 14px;
    height: 14px;
    background: white;
    border-radius: 50%;
    transition: transform 0.2s;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }

  .toggle input:checked + .toggle-track {
    background: var(--_accent);
  }

  .toggle input:checked + .toggle-track::after {
    transform: translateX(16px);
  }

  .toggle input:disabled + .toggle-track {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .cookie-toggle {
    position: absolute;
    left: 16px;
    top: 14px;
    background: none;
    border: none;
    cursor: pointer;
    padding: 2px;
    margin: 0;
    color: var(--_text-secondary);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .cookie-toggle:hover {
    color: var(--_accent);
  }

  .cookie-toggle svg {
    transition: transform 0.2s;
    transform: rotate(-90deg);
  }

  .cookie-toggle.open svg {
    transform: rotate(0deg);
  }

  probo-cookie-list {
    display: flex;
    flex-direction: column;
    margin-top: 10px;
    background: color-mix(in srgb, var(--_text) 4%, var(--_bg));
    border-radius: 8px;
    overflow: hidden;
  }

  .cookie-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 10px 12px;
    font-size: calc(var(--_font-size) - 2px);
    border-bottom: 1px solid var(--_border);
  }

  .cookie-item:last-child {
    border-bottom: none;
  }

  .cookie-name {
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 500;
    font-family: monospace;
  }

  .cookie-detail {
    font-weight: 500;
  }

  .cookie-detail > span {
    font-weight: 400;
  }

  .branding {
    text-align: center;
  }

  .branding a {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: calc(var(--_font-size) - 2px);
    font-weight: 400;
    color: var(--_text-secondary);
    text-decoration: none;
  }

  .branding a:hover {
    color: var(--_text);
  }

  .branding svg {
    flex-shrink: 0;
  }

  [hidden] {
    display: none !important;
  }
`;
