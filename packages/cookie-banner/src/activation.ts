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

function deactivateElement(el: Element): void {
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
}

function removeCookies(names: string[]): void {
  const parts = location.hostname.split(".");
  const rootDomain =
    parts.length > 1 ? "." + parts.slice(-2).join(".") : location.hostname;

  for (const name of names) {
    document.cookie = `${name}=; path=/; max-age=0`;
    document.cookie = `${name}=; path=/; domain=${rootDomain}; max-age=0`;
  }
}

export function deactivateElements(
  consentData: Record<string, boolean>,
  categoryCookies: Record<string, string[]>,
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
      deactivateElement(el);
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
