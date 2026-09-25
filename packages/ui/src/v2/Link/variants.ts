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

// Text link (Link / Anchor). Default is always underlined; `underline={false}`
// keeps plain text until hover (meta rows, icon+label chrome). Sizes align
// with Text; color × highContrast compounds mirror typography neutrals.
// Button-looking navigation uses ButtonLink / ButtonAnchor instead.
export const link = tv({
  base: [
    "inline-flex items-center underline-offset-2",
    "cursor-pointer outline-none transition-colors",
    "focus-visible:rounded-1 focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
  ],
  variants: {
    size: {
      1: "gap-1 text-1 [&_svg]:size-3.5",
      2: "gap-1.5 text-2 [&_svg]:size-4",
      3: "gap-1.5 text-3 [&_svg]:size-4",
      4: "gap-2 text-4 [&_svg]:size-5",
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
    highContrast: {
      true: "",
      false: "",
    },
    // Look-only — does not change structure. false = underline on hover.
    underline: {
      true: "underline",
      false: "no-underline hover:underline",
    },
  },
  compoundVariants: [
    { color: "neutral", highContrast: false, class: "text-sand-11 hover:text-sand-12" },
    { color: "neutral", highContrast: true, class: "text-sand-12 hover:text-sand-12" },
    { color: "gold", highContrast: false, class: "text-gold-11 hover:text-gold-12" },
    { color: "gold", highContrast: true, class: "text-gold-12 hover:text-gold-12" },
    { color: "red", highContrast: false, class: "text-red-11 hover:text-red-12" },
    { color: "red", highContrast: true, class: "text-red-12 hover:text-red-12" },
    { color: "green", highContrast: false, class: "text-green-11 hover:text-green-12" },
    { color: "green", highContrast: true, class: "text-green-12 hover:text-green-12" },
    { color: "amber", highContrast: false, class: "text-amber-11 hover:text-amber-12" },
    { color: "amber", highContrast: true, class: "text-amber-12 hover:text-amber-12" },
    { color: "sky", highContrast: false, class: "text-sky-11 hover:text-sky-12" },
    { color: "sky", highContrast: true, class: "text-sky-12 hover:text-sky-12" },
    { color: "indigo", highContrast: false, class: "text-indigo-11 hover:text-indigo-12" },
    { color: "indigo", highContrast: true, class: "text-indigo-12 hover:text-indigo-12" },
  ],
  defaultVariants: {
    size: 2,
    color: "neutral",
    highContrast: false,
    underline: true,
  },
});
