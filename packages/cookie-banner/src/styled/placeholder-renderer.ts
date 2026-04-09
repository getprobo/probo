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

import type { ThemeConfig, WidgetStrings } from "../headless/types";

const placeholderMap = new WeakMap<HTMLElement, HTMLElement>();
const originalDisplayMap = new WeakMap<HTMLElement, string>();

const LOCK_ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>`;

function createPlaceholder(
  el: HTMLElement,
  categoryName: string,
  categoryId: string,
  onAcceptCategory: (categoryId: string) => void,
  strings: WidgetStrings,
  theme: ThemeConfig,
): HTMLElement {
  const placeholder = document.createElement("div");
  placeholder.setAttribute("data-probo-placeholder", "true");

  // Copy dimensions from the original element.
  const width = el.getAttribute("width");
  const height = el.getAttribute("height");
  const computedStyle = window.getComputedStyle(el);

  const placeholderWidth = width
    ? `${width}${width.includes("%") ? "" : "px"}`
    : computedStyle.width;
  const placeholderHeight = height
    ? `${height}${height.includes("%") ? "" : "px"}`
    : computedStyle.height;

  placeholder.style.cssText = [
    `display:flex`,
    `flex-direction:column`,
    `align-items:center`,
    `justify-content:center`,
    `gap:12px`,
    `width:${placeholderWidth}`,
    `height:${placeholderHeight}`,
    `min-height:120px`,
    `box-sizing:border-box`,
    `padding:24px`,
    `background-color:${theme.background_color}`,
    `border:1px solid ${theme.border_color}`,
    `border-radius:${theme.border_radius}px`,
    `font-family:${theme.font_family}`,
    `color:${theme.text_color}`,
    `text-align:center`,
    `overflow:hidden`,
  ].join(";");

  // Icon
  const icon = document.createElement("div");
  icon.style.cssText = [
    `color:${theme.secondary_text_body_color}`,
    `opacity:0.6`,
  ].join(";");
  icon.innerHTML = LOCK_ICON_SVG;
  placeholder.appendChild(icon);

  // Message
  const message = document.createElement("p");
  message.style.cssText = [
    `margin:0`,
    `font-size:14px`,
    `line-height:1.4`,
    `color:${theme.secondary_text_body_color}`,
    `max-width:360px`,
  ].join(";");
  message.textContent = strings.contextualBlockedMessage.replace(
    "{categoryName}",
    categoryName,
  );
  placeholder.appendChild(message);

  // Button
  const button = document.createElement("button");
  button.style.cssText = [
    `display:inline-flex`,
    `align-items:center`,
    `gap:6px`,
    `padding:8px 16px`,
    `border:none`,
    `border-radius:${theme.border_radius}px`,
    `background-color:${theme.primary_color}`,
    `color:${theme.primary_text_color}`,
    `font-family:${theme.font_family}`,
    `font-size:14px`,
    `font-weight:500`,
    `cursor:pointer`,
    `transition:opacity 0.15s`,
  ].join(";");
  button.textContent = strings.contextualAllowButton;
  button.addEventListener("mouseenter", () => {
    button.style.opacity = "0.85";
  });
  button.addEventListener("mouseleave", () => {
    button.style.opacity = "1";
  });
  button.addEventListener("click", (e) => {
    e.preventDefault();
    e.stopPropagation();
    onAcceptCategory(categoryId);
  });
  placeholder.appendChild(button);

  return placeholder;
}

export function applyContextualPlaceholders(
  categories: Record<string, boolean>,
  getCategoryName: (id: string) => string,
  onAcceptCategory: (categoryId: string) => void,
  strings: WidgetStrings,
  theme: ThemeConfig,
): void {
  try {
    const elements = document.querySelectorAll(
      "iframe[data-cookie-category], img[data-cookie-category]",
    );

    elements.forEach((node) => {
      const el = node as HTMLElement;
      const categoryId = el.getAttribute("data-cookie-category");
      if (!categoryId) return;

      const consented = categories[categoryId] === true;

      if (!consented) {
        // Already has a placeholder — skip.
        if (placeholderMap.has(el)) return;

        const categoryName = getCategoryName(categoryId);
        const placeholder = createPlaceholder(
          el,
          categoryName,
          categoryId,
          onAcceptCategory,
          strings,
          theme,
        );

        // Save original display value and hide the element.
        originalDisplayMap.set(el, el.style.display);
        el.style.display = "none";

        // Insert placeholder right before the hidden element.
        el.parentNode?.insertBefore(placeholder, el);
        placeholderMap.set(el, placeholder);
      } else {
        // Consented — remove placeholder if present.
        const placeholder = placeholderMap.get(el);
        if (placeholder) {
          placeholder.remove();
          placeholderMap.delete(el);
          el.style.display = originalDisplayMap.get(el) ?? "";
          originalDisplayMap.delete(el);
        }
      }
    });
  } catch {
    // Never break the host site.
  }
}

export function removeAllPlaceholders(): void {
  try {
    const placeholders = document.querySelectorAll(
      "[data-probo-placeholder]",
    );
    placeholders.forEach((placeholder) => {
      const el = placeholder.nextElementSibling as HTMLElement | null;
      if (el) {
        el.style.display = originalDisplayMap.get(el) ?? "";
        originalDisplayMap.delete(el);
        placeholderMap.delete(el);
      }
      placeholder.remove();
    });
  } catch {
    // Never break the host site.
  }
}
