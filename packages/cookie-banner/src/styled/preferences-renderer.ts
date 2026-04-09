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

import type { BannerCategory, ThemeConfig, WidgetStrings } from "../headless/types";

export function renderPreferences(
  container: HTMLElement,
  categories: BannerCategory[],
  currentConsent: Record<string, boolean>,
  onSave: (choices: Record<string, boolean>) => void,
  onRejectAll: () => void,
  onAcceptAll: () => void,
  strings: WidgetStrings,
  theme: ThemeConfig,
): void {
  const panel = document.createElement("div");
  panel.style.cssText =
    "padding:24px;max-height:80vh;overflow-y:auto;";
  panel.setAttribute("role", "region");
  panel.setAttribute("aria-label", strings.cookiePreferences);

  const title = document.createElement("h2");
  title.textContent = strings.cookiePreferences;
  title.style.cssText =
    `margin:0 0 16px;font-size:18px;font-weight:600;color:${theme.text_color};`;
  panel.appendChild(title);

  const checkboxes: { id: string; checkbox: HTMLInputElement }[] = [];

  for (const cat of categories) {
    const section = document.createElement("div");
    section.style.cssText =
      `padding:12px 0;border-bottom:1px solid ${theme.border_color};`;

    const row = document.createElement("div");
    row.style.cssText =
      "display:flex;justify-content:space-between;align-items:flex-start;";

    const info = document.createElement("div");
    info.style.cssText = "flex:1;margin-right:16px;";

    const nameRow = document.createElement("div");
    nameRow.style.cssText = "display:flex;align-items:center;gap:6px;";

    if (cat.cookies.length > 0) {
      const arrow = document.createElement("span");
      arrow.textContent = "\u25B6";
      arrow.style.cssText =
        `font-size:10px;color:${theme.secondary_text_body_color};cursor:pointer;transition:transform .2s;user-select:none;`;
      nameRow.appendChild(arrow);

      const cookieDetails = document.createElement("div");
      cookieDetails.style.cssText = "display:none;margin-top:8px;";

      const toggle = () => {
        const isOpen = cookieDetails.style.display !== "none";
        cookieDetails.style.display = isOpen ? "none" : "block";
        arrow.style.transform = isOpen ? "rotate(0deg)" : "rotate(90deg)";
      };

      nameRow.style.cursor = "pointer";
      nameRow.addEventListener("click", toggle);

      renderCookieDetails(cookieDetails, cat, theme);
      info.appendChild(nameRow);

      if (cat.description) {
        const desc = document.createElement("div");
        desc.textContent = cat.description;
        desc.style.cssText =
          `font-size:12px;color:${theme.secondary_text_body_color};margin-top:4px;line-height:1.4;`;
        info.appendChild(desc);
      }

      info.appendChild(cookieDetails);
    } else {
      info.appendChild(nameRow);

      if (cat.description) {
        const desc = document.createElement("div");
        desc.textContent = cat.description;
        desc.style.cssText =
          `font-size:12px;color:${theme.secondary_text_body_color};margin-top:4px;line-height:1.4;`;
        info.appendChild(desc);
      }
    }

    const name = document.createElement("span");
    name.textContent = cat.name;
    name.style.cssText = `font-weight:500;font-size:14px;color:${theme.text_color};`;
    nameRow.appendChild(name);

    row.appendChild(info);

    const toggleLabel = document.createElement("label");
    toggleLabel.style.cssText =
      `position:relative;display:inline-block;width:44px;height:24px;flex-shrink:0;${cat.required ? "opacity:0.5;" : ""}`;

    const checkbox = document.createElement("input");
    checkbox.type = "checkbox";
    checkbox.checked = cat.required || (currentConsent[cat.id] ?? false);
    checkbox.disabled = cat.required;
    checkbox.style.cssText = "opacity:0;width:0;height:0;";
    checkbox.setAttribute("aria-label", cat.name);
    checkbox.setAttribute("role", "switch");
    checkbox.setAttribute("aria-checked", String(checkbox.checked));

    const slider = document.createElement("span");
    const updateSlider = () => {
      slider.style.cssText = `position:absolute;cursor:${cat.required ? "not-allowed" : "pointer"};top:0;left:0;right:0;bottom:0;background:${checkbox.checked ? theme.primary_color : "#d1d5db"};border-radius:12px;transition:background .2s;`;
      const knob = slider.querySelector("span") as HTMLSpanElement;
      if (knob) {
        knob.style.transform = checkbox.checked
          ? "translateX(20px)"
          : "translateX(0)";
      }
    };

    const knob = document.createElement("span");
    knob.style.cssText =
      "position:absolute;height:18px;width:18px;left:3px;bottom:3px;background:white;border-radius:50%;transition:transform .2s;";
    slider.appendChild(knob);

    toggleLabel.appendChild(checkbox);
    toggleLabel.appendChild(slider);
    row.appendChild(toggleLabel);
    section.appendChild(row);
    panel.appendChild(section);

    updateSlider();
    checkbox.addEventListener("change", () => {
      updateSlider();
      checkbox.setAttribute("aria-checked", String(checkbox.checked));
    });

    checkboxes.push({ id: cat.id, checkbox });
  }

  const actions = document.createElement("div");
  actions.style.cssText =
    "display:flex;gap:12px;margin-top:20px;";

  const rejectBtn = document.createElement("button");
  rejectBtn.textContent = strings.rejectAll;
  rejectBtn.style.cssText =
    `flex:1;padding:12px;border:1px solid ${theme.border_color};border-radius:${theme.border_radius}px;background:${theme.background_color};color:${theme.text_color};font-size:14px;font-weight:500;cursor:pointer;font-family:${theme.font_family};`;
  rejectBtn.addEventListener("click", onRejectAll);
  actions.appendChild(rejectBtn);

  const saveBtn = document.createElement("button");
  saveBtn.textContent = strings.savePreferences;
  saveBtn.style.cssText =
    `flex:1;padding:12px;border:1px solid ${theme.border_color};border-radius:${theme.border_radius}px;background:${theme.background_color};color:${theme.text_color};font-size:14px;font-weight:500;cursor:pointer;font-family:${theme.font_family};`;
  saveBtn.addEventListener("click", () => {
    const choices: Record<string, boolean> = {};
    for (const { id, checkbox } of checkboxes) {
      choices[id] = checkbox.checked;
    }
    onSave(choices);
  });
  actions.appendChild(saveBtn);

  const acceptBtn = document.createElement("button");
  acceptBtn.textContent = strings.acceptAll;
  acceptBtn.style.cssText =
    `flex:1;padding:12px;border:none;border-radius:${theme.border_radius}px;background:${theme.primary_color};color:${theme.primary_text_color};font-size:14px;font-weight:500;cursor:pointer;font-family:${theme.font_family};`;
  acceptBtn.addEventListener("click", onAcceptAll);
  actions.appendChild(acceptBtn);

  panel.appendChild(actions);
  container.appendChild(panel);
}

function renderCookieDetails(
  container: HTMLElement,
  cat: BannerCategory,
  theme: ThemeConfig,
): void {
  for (const cookie of cat.cookies) {
    const item = document.createElement("div");
    item.style.cssText =
      `padding:8px 0;border-bottom:1px solid ${theme.border_color};`;

    const fields: [string, string][] = [
      ["Cookie", cookie.name],
      ["Duration", cookie.duration],
      ["Description", cookie.description],
    ];

    for (const [label, value] of fields) {
      if (!value) continue;
      const fieldRow = document.createElement("div");
      fieldRow.style.cssText =
        "display:flex;gap:12px;padding:2px 0;font-size:12px;line-height:1.4;";

      const labelEl = document.createElement("span");
      labelEl.textContent = label;
      labelEl.style.cssText =
        `font-weight:600;min-width:80px;flex-shrink:0;color:${theme.text_color};`;
      fieldRow.appendChild(labelEl);

      const valueEl = document.createElement("span");
      valueEl.textContent = value;
      valueEl.style.cssText =
        `color:${theme.secondary_text_body_color};`;
      fieldRow.appendChild(valueEl);

      item.appendChild(fieldRow);
    }

    container.appendChild(item);
  }
}
