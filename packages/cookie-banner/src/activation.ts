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

import type { BannerTexts } from "./i18n";
import { removeCookies } from "./cookie-utils";
import { LOCK_ICON } from "./html";

const ATTR_CATEGORY = "data-cookie-consent";
const ATTR_SRC = "data-src";
const ATTR_HREF = "data-href";
const ATTR_ACTIVATED = "data-cookie-consent-activated";
const ATTR_PLACEHOLDER = "data-cookie-consent-placeholder";
const ATTR_HIDDEN = "data-cookie-consent-hidden";

const ACTIVATABLE_TAGS = new Set([
  "SCRIPT",
  "IFRAME",
  "IMG",
  "VIDEO",
  "AUDIO",
  "EMBED",
  "OBJECT",
  "LINK",
]);

const VISUAL_TAGS = new Set([
  "IFRAME",
  "IMG",
  "VIDEO",
  "AUDIO",
  "EMBED",
  "OBJECT",
]);


const PLACEHOLDER_STYLES = `
[${ATTR_HIDDEN}] {
  display: none !important;
}
[${ATTR_PLACEHOLDER}] {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 12px;
  padding: 24px;
  background: var(--probo-bg, #ffffff);
  color: var(--probo-text-secondary, #555555);
  border: 1px dashed var(--probo-border, #e0e0e0);
  border-radius: var(--probo-radius, 12px);
  font-family: var(--probo-font-family, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif);
  font-size: var(--probo-font-size, 14px);
  line-height: 1.5;
  text-align: center;
  box-sizing: border-box;
  min-height: 120px;
}
[${ATTR_PLACEHOLDER}] .probo-ph-icon {
  color: var(--probo-text-secondary, #555555);
}
[${ATTR_PLACEHOLDER}] .probo-ph-text {
  margin: 0;
}
[${ATTR_PLACEHOLDER}] .probo-ph-text strong {
  font-weight: 600;
}
[${ATTR_PLACEHOLDER}] .probo-ph-link {
  background: none;
  border: none;
  color: var(--probo-accent, #1a1a1a);
  text-decoration: underline;
  cursor: pointer;
  font-family: inherit;
  font-size: inherit;
  padding: 0;
}
[${ATTR_PLACEHOLDER}] .probo-ph-link:hover {
  opacity: 0.8;
}
`;

let stylesInjected = false;

function injectPlaceholderStyles(): void {
  if (stylesInjected) return;
  stylesInjected = true;

  const style = document.createElement("style");
  style.id = "probo-placeholder-styles";
  style.textContent = PLACEHOLDER_STYLES;
  document.head.appendChild(style);
}

function escapeHtml(s: string): string {
  const el = document.createElement("span");
  el.textContent = s;
  return el.innerHTML;
}

function createPlaceholder(
  el: Element,
  category: string,
  label?: string,
  texts?: BannerTexts,
): void {
  if (el.hasAttribute(ATTR_HIDDEN)) return;

  injectPlaceholderStyles();

  const displayLabel = label || category;

  const placeholder = document.createElement("div");
  placeholder.setAttribute(ATTR_PLACEHOLDER, category);

  const htmlEl = el as HTMLElement;

  if (htmlEl.className) placeholder.className = htmlEl.className;

  const w = el.getAttribute("width");
  if (w) placeholder.style.width = /^\d+$/.test(w) ? w + "px" : w;

  const h = el.getAttribute("height");
  if (h) placeholder.style.height = /^\d+$/.test(h) ? h + "px" : h;

  const DIMENSIONAL_PROPS = [
    "width",
    "height",
    "min-width",
    "min-height",
    "max-width",
    "max-height",
    "aspect-ratio",
    "margin",
    "margin-top",
    "margin-right",
    "margin-bottom",
    "margin-left",
  ];
  if (htmlEl.style) {
    for (const prop of DIMENSIONAL_PROPS) {
      const val = htmlEl.style.getPropertyValue(prop);
      if (val) placeholder.style.setProperty(prop, val);
    }
  }

  let phText: string;
  if (texts?.placeholder_text) {
    const parts = texts.placeholder_text.split("{{category}}");
    phText = parts.map(p => escapeHtml(p)).join(`<strong>${escapeHtml(displayLabel)}</strong>`);
  } else {
    phText = `This content requires <strong>${escapeHtml(displayLabel)}</strong> cookies.`;
  }
  const phButton = texts?.placeholder_button ?? "Manage cookie preferences";

  placeholder.innerHTML = [
    `<span class="probo-ph-icon">${LOCK_ICON}</span>`,
    `<p class="probo-ph-text">${phText}</p>`,
    `<button type="button" class="probo-ph-link">${escapeHtml(phButton)}</button>`,
  ].join("");

  placeholder.querySelector(".probo-ph-link")!.addEventListener("click", () => {
    document.dispatchEvent(new CustomEvent("probo-open-preferences"));
  });

  el.setAttribute(ATTR_HIDDEN, "");
  el.parentNode!.insertBefore(placeholder, el.nextSibling);
}

function removePlaceholder(el: Element): void {
  if (!el.hasAttribute(ATTR_HIDDEN)) return;

  const next = el.nextElementSibling;
  if (next?.hasAttribute(ATTR_PLACEHOLDER)) {
    next.remove();
  }

  el.removeAttribute(ATTR_HIDDEN);
}

function activateScript(el: HTMLScriptElement): void {
  const replacement = document.createElement("script");
  const originalType = el.getAttribute("data-type");
  const category = el.getAttribute(ATTR_CATEGORY);

  for (const attr of el.attributes) {
    if (
      attr.name === "type" ||
      attr.name === "data-type" ||
      attr.name === ATTR_CATEGORY
    ) {
      continue;
    }
    if (attr.name === ATTR_SRC) {
      replacement.setAttribute("src", attr.value);
      continue;
    }
    replacement.setAttribute(attr.name, attr.value);
  }

  if (originalType) {
    replacement.setAttribute("type", originalType);
  }
  replacement.setAttribute(ATTR_ACTIVATED, category || "");

  if (el.textContent) {
    replacement.textContent = el.textContent;
  }

  el.parentNode!.replaceChild(replacement, el);
}

function activateElement(el: Element): void {
  removePlaceholder(el);

  const category = el.getAttribute(ATTR_CATEGORY);

  const src = el.getAttribute(ATTR_SRC);
  if (src) {
    el.setAttribute("src", src);
    el.removeAttribute(ATTR_SRC);
  }

  const href = el.getAttribute(ATTR_HREF);
  if (href) {
    el.setAttribute("href", href);
    el.removeAttribute(ATTR_HREF);
  }

  el.removeAttribute(ATTR_CATEGORY);
  el.setAttribute(ATTR_ACTIVATED, category || "");
}

function tryActivate(
  el: Element,
  consentData: Record<string, boolean>,
): void {
  if (el.hasAttribute(ATTR_ACTIVATED)) {
    return;
  }

  if (!ACTIVATABLE_TAGS.has(el.tagName)) {
    return;
  }

  const category = el.getAttribute(ATTR_CATEGORY);
  if (!category || !consentData[category]) {
    return;
  }

  if (el instanceof HTMLScriptElement) {
    activateScript(el);
  } else {
    activateElement(el);
  }
}

function deactivateScript(el: HTMLScriptElement): void {
  const category = el.getAttribute(ATTR_ACTIVATED);
  const currentType = el.getAttribute("type");
  const replacement = document.createElement("script");

  for (const attr of el.attributes) {
    if (
      attr.name === ATTR_ACTIVATED ||
      attr.name === "type" ||
      attr.name === "src"
    ) {
      continue;
    }
    replacement.setAttribute(attr.name, attr.value);
  }

  const src = el.getAttribute("src");
  if (src) {
    replacement.setAttribute(ATTR_SRC, src);
  }

  if (currentType) {
    replacement.setAttribute("data-type", currentType);
  }
  replacement.setAttribute("type", "text/plain");

  if (category) {
    replacement.setAttribute(ATTR_CATEGORY, category);
  }

  if (el.textContent) {
    replacement.textContent = el.textContent;
  }

  el.parentNode!.replaceChild(replacement, el);
}

function deactivateElement(el: Element, label?: string): void {
  const category = el.getAttribute(ATTR_ACTIVATED);

  const src = el.getAttribute("src");
  if (src) {
    el.setAttribute(ATTR_SRC, src);
    el.removeAttribute("src");
  }

  const href = el.getAttribute("href");
  if (href) {
    el.setAttribute(ATTR_HREF, href);
    el.removeAttribute("href");
  }

  if (category) {
    el.setAttribute(ATTR_CATEGORY, category);
  }
  el.removeAttribute(ATTR_ACTIVATED);

  if (category && VISUAL_TAGS.has(el.tagName)) {
    createPlaceholder(el, category, label);
  }
}


export function deactivateElements(
  consentData: Record<string, boolean>,
  categoryCookies: Record<string, string[]>,
  categoryLabels: Record<string, string>,
): void {
  const elements = document.querySelectorAll(`[${ATTR_ACTIVATED}]`);
  const cookiesToRemove = new Set<string>();

  for (const el of elements) {
    const category = el.getAttribute(ATTR_ACTIVATED);
    if (!category || consentData[category]) {
      continue;
    }

    if (el instanceof HTMLScriptElement) {
      deactivateScript(el);
    } else {
      deactivateElement(el, categoryLabels[category]);
    }

    const cookies = categoryCookies[category];
    if (cookies) {
      for (const name of cookies) {
        cookiesToRemove.add(name);
      }
    }
  }

  if (cookiesToRemove.size > 0) {
    removeCookies([...cookiesToRemove]);
  }
}

export function activateElements(
  consentData: Record<string, boolean>,
): void {
  const elements = document.querySelectorAll(`[${ATTR_CATEGORY}]`);
  for (const el of elements) {
    tryActivate(el, consentData);
  }
}

export function addPlaceholders(
  consentData: Record<string, boolean>,
  categoryLabels: Record<string, string>,
  texts?: BannerTexts,
): void {
  const elements = document.querySelectorAll(`[${ATTR_CATEGORY}]`);
  for (const el of elements) {
    const category = el.getAttribute(ATTR_CATEGORY);
    if (!category || consentData[category]) continue;
    if (!VISUAL_TAGS.has(el.tagName)) continue;
    createPlaceholder(el, category, categoryLabels[category], texts);
  }
}

function tryPlaceholder(
  el: Element,
  consentData: Record<string, boolean>,
  categoryLabels: Record<string, string>,
  texts?: BannerTexts,
): void {
  const category = el.getAttribute(ATTR_CATEGORY);
  if (!category || consentData[category]) return;
  if (!VISUAL_TAGS.has(el.tagName)) return;
  createPlaceholder(el, category, categoryLabels[category], texts);
}

export function observeAndActivate(
  consentData: Record<string, boolean>,
  categoryLabels: Record<string, string>,
  texts?: BannerTexts,
): MutationObserver {
  const observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      for (const node of mutation.addedNodes) {
        if (!(node instanceof Element)) {
          continue;
        }

        if (node.hasAttribute(ATTR_CATEGORY)) {
          tryActivate(node, consentData);
          tryPlaceholder(node, consentData, categoryLabels, texts);
        }

        const nested = node.querySelectorAll(`[${ATTR_CATEGORY}]`);
        for (const el of nested) {
          tryActivate(el, consentData);
          tryPlaceholder(el, consentData, categoryLabels, texts);
        }
      }
    }
  });

  observer.observe(document.documentElement, {
    childList: true,
    subtree: true,
  });

  return observer;
}
