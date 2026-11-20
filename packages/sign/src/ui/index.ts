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

import { ProboDocuments } from "../client";
import type { AcceptanceResult, ProboSignOptions } from "../types";

export interface MountCheckboxOptions extends ProboSignOptions {
  /** Override the label text. Defaults to the document's consent_text. */
  label?: string;
  /** Called with the acceptance result after a successful accept(). */
  onAccepted?: (results: AcceptanceResult[]) => void;
}

export interface CheckboxHandle {
  /** Remove the rendered checkbox and clean up. */
  unmount(): void;
}

/**
 * Renders a consent checkbox into `container`. The label text defaults to the
 * document's canonical consent_text. A "View document" link is shown when the
 * document exposes a PDF URL.
 *
 * The checkbox starts unchecked. On check it calls accept() and fires onAccepted.
 * On API error it unchecks itself and re-throws.
 */
export function mountCheckbox(
  container: HTMLElement,
  options: MountCheckboxOptions,
): CheckboxHandle {
  const client = new ProboDocuments(options);
  const { publishableKey, label, onAccepted } = options;

  const wrapper = document.createElement("div");
  wrapper.setAttribute("data-probo-sign", "checkbox");
  wrapper.style.cssText = "display:flex;align-items:flex-start;gap:8px;font-family:inherit";

  const checkbox = document.createElement("input");
  checkbox.type = "checkbox";
  checkbox.disabled = true;
  checkbox.id = `probo-sign-${Math.random().toString(36).slice(2)}`;
  checkbox.style.cssText = "margin-top:3px;flex-shrink:0;cursor:pointer";

  const labelEl = document.createElement("label");
  labelEl.htmlFor = checkbox.id;
  labelEl.style.cssText = "cursor:pointer;line-height:1.4";
  labelEl.textContent = label ?? "Loading…";

  wrapper.appendChild(checkbox);
  wrapper.appendChild(labelEl);
  container.appendChild(wrapper);

  let viewLink: HTMLAnchorElement | null = null;

  client
    .loadDocument(publishableKey)
    .then((doc) => {
      labelEl.textContent = label ?? doc.consentText;
      checkbox.disabled = false;

      if (doc.pdfUrl) {
        viewLink = document.createElement("a");
        viewLink.href = doc.pdfUrl;
        viewLink.target = "_blank";
        viewLink.rel = "noopener noreferrer";
        viewLink.textContent = " View document";
        viewLink.style.cssText = "white-space:nowrap";
        labelEl.appendChild(viewLink);
      }
    })
    .catch(() => {
      labelEl.textContent = "Document unavailable";
    });

  checkbox.addEventListener("change", () => {
    if (!checkbox.checked) return;
    checkbox.disabled = true;

    client
      .accept([publishableKey])
      .then((results) => {
        onAccepted?.(results);
      })
      .catch((err: unknown) => {
        checkbox.checked = false;
        checkbox.disabled = false;
        throw err;
      });
  });

  return {
    unmount(): void {
      container.removeChild(wrapper);
    },
  };
}
