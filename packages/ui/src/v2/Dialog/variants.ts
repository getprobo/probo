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

import { tv } from "tailwind-variants/lite";

// Modal dialog (Radix "Dialog" over Base UI). `placement` pins the popup to the
// center or near the top. The popup gets vertical padding + a 16px gap; each
// region (header, body, footer) carries its own horizontal padding so a
// full-bleed body slot stays possible.
export const dialog = tv({
  slots: {
    backdrop: [
      "fixed inset-0 z-4 bg-sand-12/40",
      "transition-opacity duration-150",
      "data-starting-style:opacity-0 data-ending-style:opacity-0",
    ],
    popup: [
      "fixed left-1/2 z-5 -translate-x-1/2",
      "flex w-[calc(100vw-2rem)] max-w-150 flex-col gap-4",
      "max-h-[calc(100vh-2rem)] overflow-x-clip",
      "rounded-5 border border-sand-6 bg-sand-1 py-6 shadow-6 outline-none",
      "transition-all duration-150",
      "data-starting-style:scale-95 data-starting-style:opacity-0",
      "data-ending-style:scale-95 data-ending-style:opacity-0",
    ],
    header: "flex flex-col gap-2 px-6",
    title: "text-4 font-medium text-sand-12",
    description: "text-2 text-sand-11",
    body: "px-6",
    footer: "flex flex-wrap items-center justify-end gap-3 px-6 max-sm:flex-col-reverse max-sm:items-stretch",
    // Fixed to the viewport so absolute menus inside it do not grow the page
    // and shift the dialog. Clicks pass through the empty layer.
    overlayRoot: "pointer-events-none fixed inset-0 z-5 [&>*]:pointer-events-auto",
  },
  variants: {
    lockScroll: {
      true: { popup: "overflow-y-hidden" },
      false: { popup: "overflow-y-auto" },
    },
    placement: {
      center: { popup: "top-1/2 -translate-y-1/2" },
      top: { popup: "top-24" },
    },
  },
  defaultVariants: {
    lockScroll: false,
    placement: "center",
  },
});

// Static frame matching the dialog popup, without the interactive positioning,
// so a placeholder can render before Base UI (and the content) loads.
export const dialogSkeleton = tv({
  base: [
    "flex w-full max-w-150 flex-col gap-4",
    "rounded-5 border border-sand-6 bg-sand-1 py-6 shadow-6",
  ],
});

// `chip` adds the filled keycap. Neutral chips sit one step lighter than the
// button, with a border one step darker. Colored chips sit one step darker
// than the button, and the border stays one step darker than that fill.
// Hint text uses the button's text color. Soft buttons are step 3. Solid
// buttons are step 9.
export const shortcutHint = tv({
  slots: {
    hint: [
      "pointer-events-none inline-flex shrink-0 items-center gap-px",
      "text-1 leading-none whitespace-nowrap text-current select-none",
    ],
    modifier: "size-2.5 shrink-0",
  },
  variants: {
    chip: {
      true: {
        hint: "rounded-1 border px-0.5 py-px",
      },
      false: {},
    },
    variant: {
      soft: "",
      solid: "",
    },
    color: {
      neutral: "",
      gold: "",
      red: "",
      green: "",
      amber: "",
      sky: "",
      indigo: "",
    },
  },
  compoundVariants: [
    { chip: true, variant: "soft", color: "neutral", class: { hint: "border-sand-4 bg-sand-2" } },
    { chip: true, variant: "soft", color: "gold", class: { hint: "border-gold-5 bg-gold-4" } },
    { chip: true, variant: "soft", color: "red", class: { hint: "border-red-5 bg-red-4" } },
    { chip: true, variant: "soft", color: "green", class: { hint: "border-green-5 bg-green-4" } },
    { chip: true, variant: "soft", color: "amber", class: { hint: "border-amber-5 bg-amber-4" } },
    { chip: true, variant: "soft", color: "sky", class: { hint: "border-sky-5 bg-sky-4" } },
    { chip: true, variant: "soft", color: "indigo", class: { hint: "border-indigo-5 bg-indigo-4" } },
    { chip: true, variant: "solid", color: "neutral", class: { hint: "border-sand-10 bg-sand-8" } },
    { chip: true, variant: "solid", color: "gold", class: { hint: "border-gold-11 bg-gold-10" } },
    { chip: true, variant: "solid", color: "red", class: { hint: "border-red-11 bg-red-10" } },
    { chip: true, variant: "solid", color: "green", class: { hint: "border-green-11 bg-green-10" } },
    { chip: true, variant: "solid", color: "amber", class: { hint: "border-amber-11 bg-amber-10" } },
    { chip: true, variant: "solid", color: "sky", class: { hint: "border-sky-11 bg-sky-10" } },
    { chip: true, variant: "solid", color: "indigo", class: { hint: "border-indigo-11 bg-indigo-10" } },
  ],
  defaultVariants: {
    chip: false,
    variant: "soft",
    color: "neutral",
  },
});
