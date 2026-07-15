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

// Styled Base UI toast. Each toast is an elevated card that reuses Callout's
// `surface` hue language (tinted background + hue border, icon and body in the
// hue text step, title one step higher for contrast) so toasts and callouts
// read as one family. Meaning is carried by the icon too, not color alone.
export const toaster = tv({
  slots: {
    viewport: [
      "fixed bottom-0 right-0 z-60 flex w-[380px] max-w-[calc(100vw-2rem)] flex-col gap-2 p-4",
      "outline-none",
    ],
    root: [
      "flex items-start gap-3 rounded-4 border p-4 shadow-4",
      "transition-all duration-150",
      "data-starting-style:translate-y-2 data-starting-style:opacity-0",
      "data-ending-style:translate-y-2 data-ending-style:opacity-0",
    ],
    // Icon and description inherit the root's hue text step (currentColor).
    icon: "mt-px shrink-0 [&_svg]:size-5",
    content: "flex min-w-0 flex-1 flex-col gap-1",
    title: "text-2 font-medium",
    description: "text-1",
    close: "-mr-1 -mt-1 shrink-0 rounded-2 p-1 opacity-70 transition-opacity hover:opacity-100 [&_svg]:size-4",
  },
  variants: {
    // Mirrors Callout's surface tokens: bg step 2, border step 6, text step 11,
    // with the title bumped to step 12.
    type: {
      neutral: {},
      success: {},
      error: {},
      warning: {},
      info: {},
    },
  },
  compoundVariants: [
    { type: "neutral", class: { root: "bg-sand-2 border-sand-6 text-sand-11", title: "text-sand-12" } },
    { type: "success", class: { root: "bg-green-2 border-green-6 text-green-11", title: "text-green-12" } },
    { type: "error", class: { root: "bg-red-2 border-red-6 text-red-11", title: "text-red-12" } },
    { type: "warning", class: { root: "bg-amber-2 border-amber-6 text-amber-11", title: "text-amber-12" } },
    { type: "info", class: { root: "bg-sky-2 border-sky-6 text-sky-11", title: "text-sky-12" } },
  ],
  defaultVariants: {
    type: "neutral",
  },
});
