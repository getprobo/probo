// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

export function applyConsent(
  categories: Record<string, boolean>,
  previousCategories?: Record<string, boolean>,
): void {
  // Scripts: activate blocked scripts (cannot un-execute)
  try {
    const blocked = document.querySelectorAll(
      'script[type="text/plain"][data-cookie-category]',
    );
    blocked.forEach((el) => {
      const script = el as HTMLScriptElement;
      const category = script.getAttribute("data-cookie-category");
      if (category && categories[category]) {
        const newScript = document.createElement("script");
        for (const attr of Array.from(script.attributes)) {
          if (attr.name === "type" || attr.name === "data-cookie-category")
            continue;
          newScript.setAttribute(attr.name, attr.value);
        }
        if (script.textContent) {
          newScript.textContent = script.textContent;
        }
        script.parentNode?.replaceChild(newScript, script);
      }
    });
  } catch {
    // Never break the host site.
  }

  // Iframes: activate or deactivate
  try {
    const iframes = document.querySelectorAll(
      "iframe[data-cookie-category]",
    );
    iframes.forEach((el) => {
      const iframe = el as HTMLIFrameElement;
      const category = iframe.getAttribute("data-cookie-category");
      if (!category) return;

      if (categories[category]) {
        const dataSrc = iframe.getAttribute("data-src");
        if (dataSrc) {
          iframe.setAttribute("src", dataSrc);
          iframe.removeAttribute("data-src");
        }
      } else if (previousCategories && previousCategories[category]) {
        const src = iframe.getAttribute("src");
        if (src && src !== "about:blank") {
          iframe.setAttribute("data-src", src);
        }
        iframe.setAttribute("src", "about:blank");
      }
    });
  } catch {
    // Never break the host site.
  }

  // Images: activate or deactivate
  try {
    const images = document.querySelectorAll(
      "img[data-cookie-category]",
    );
    images.forEach((el) => {
      const img = el as HTMLImageElement;
      const category = img.getAttribute("data-cookie-category");
      if (!category) return;

      if (categories[category]) {
        const dataSrc = img.getAttribute("data-src");
        if (dataSrc) {
          img.setAttribute("src", dataSrc);
          img.removeAttribute("data-src");
        }
      } else if (previousCategories && previousCategories[category]) {
        const src = img.getAttribute("src");
        if (src) {
          img.setAttribute("data-src", src);
        }
        img.removeAttribute("src");
      }
    });
  } catch {
    // Never break the host site.
  }

  // Links (stylesheets): activate or deactivate
  try {
    const links = document.querySelectorAll(
      "link[data-cookie-category]",
    );
    links.forEach((el) => {
      const link = el as HTMLLinkElement;
      const category = link.getAttribute("data-cookie-category");
      if (!category) return;

      if (categories[category]) {
        const dataHref = link.getAttribute("data-href");
        if (dataHref) {
          link.setAttribute("href", dataHref);
          link.removeAttribute("data-href");
        }
      } else if (previousCategories && previousCategories[category]) {
        const href = link.getAttribute("href");
        if (href) {
          link.setAttribute("data-href", href);
        }
        link.removeAttribute("href");
      }
    });
  } catch {
    // Never break the host site.
  }
}
