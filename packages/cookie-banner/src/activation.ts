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

const ATTR_CATEGORY = "data-cookie-consent";
const ATTR_SRC = "data-src";
const ATTR_HREF = "data-href";
const ATTR_ACTIVATED = "data-cookie-consent-activated";

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

function activateScript(el: HTMLScriptElement): void {
  const replacement = document.createElement("script");
  const originalType = el.getAttribute("data-type");

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
  replacement.setAttribute(ATTR_ACTIVATED, "");

  if (el.textContent) {
    replacement.textContent = el.textContent;
  }

  el.parentNode!.replaceChild(replacement, el);
}

function activateElement(el: Element): void {
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
  el.setAttribute(ATTR_ACTIVATED, "");
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

export function activateElements(
  consentData: Record<string, boolean>,
): void {
  const elements = document.querySelectorAll(`[${ATTR_CATEGORY}]`);
  for (const el of elements) {
    tryActivate(el, consentData);
  }
}

export function observeAndActivate(
  consentData: Record<string, boolean>,
): MutationObserver {
  const observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      for (const node of mutation.addedNodes) {
        if (!(node instanceof Element)) {
          continue;
        }

        if (node.hasAttribute(ATTR_CATEGORY)) {
          tryActivate(node, consentData);
        }

        const nested = node.querySelectorAll(`[${ATTR_CATEGORY}]`);
        for (const el of nested) {
          tryActivate(el, consentData);
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
