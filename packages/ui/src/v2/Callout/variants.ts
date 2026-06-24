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

// Short contextual message with a leading icon (Radix "Callout").
//   root  the container surface (background/border + hue text color)
//   icon  leading icon, inherits the hue text color
//   text  the message body
export const callout = tv({
  slots: {
    root: "flex items-start rounded-3 border border-transparent",
    icon: "shrink-0",
    text: "min-w-0 flex-1",
  },
  variants: {
    size: {
      1: { root: "gap-2 p-2 text-1", icon: "[&_svg]:size-4" },
      2: { root: "gap-3 p-3 text-2", icon: "[&_svg]:size-5" },
      3: { root: "gap-3 p-4 text-3", icon: "[&_svg]:size-6" },
    },
    variant: {
      soft: {},
      surface: {},
      outline: {},
    },
    color: {
      neutral: {},
      gold: {},
      red: {},
      green: {},
      amber: {},
      sky: {},
    },
    highContrast: {
      true: {},
      false: {},
    },
  },
  compoundVariants: [
    // soft: tinted background
    { variant: "soft", color: "neutral", class: { root: "bg-sand-3 text-sand-11" } },
    { variant: "soft", color: "gold", class: { root: "bg-gold-3 text-gold-11" } },
    { variant: "soft", color: "red", class: { root: "bg-red-3 text-red-11" } },
    { variant: "soft", color: "green", class: { root: "bg-green-3 text-green-11" } },
    { variant: "soft", color: "amber", class: { root: "bg-amber-3 text-amber-11" } },
    { variant: "soft", color: "sky", class: { root: "bg-sky-3 text-sky-11" } },

    // surface: subtle background + border
    { variant: "surface", color: "neutral", class: { root: "bg-sand-2 border-sand-6 text-sand-11" } },
    { variant: "surface", color: "gold", class: { root: "bg-gold-2 border-gold-6 text-gold-11" } },
    { variant: "surface", color: "red", class: { root: "bg-red-2 border-red-6 text-red-11" } },
    { variant: "surface", color: "green", class: { root: "bg-green-2 border-green-6 text-green-11" } },
    { variant: "surface", color: "amber", class: { root: "bg-amber-2 border-amber-6 text-amber-11" } },
    { variant: "surface", color: "sky", class: { root: "bg-sky-2 border-sky-6 text-sky-11" } },

    // outline: border only
    { variant: "outline", color: "neutral", class: { root: "border-sand-6 text-sand-11" } },
    { variant: "outline", color: "gold", class: { root: "border-gold-6 text-gold-11" } },
    { variant: "outline", color: "red", class: { root: "border-red-6 text-red-11" } },
    { variant: "outline", color: "green", class: { root: "border-green-6 text-green-11" } },
    { variant: "outline", color: "amber", class: { root: "border-amber-6 text-amber-11" } },
    { variant: "outline", color: "sky", class: { root: "border-sky-6 text-sky-11" } },

    // high-contrast text (icon + body inherit the root color)
    { color: "neutral", highContrast: true, class: { root: "text-sand-12" } },
    { color: "gold", highContrast: true, class: { root: "text-gold-12" } },
    { color: "red", highContrast: true, class: { root: "text-red-12" } },
    { color: "green", highContrast: true, class: { root: "text-green-12" } },
    { color: "amber", highContrast: true, class: { root: "text-amber-12" } },
    { color: "sky", highContrast: true, class: { root: "text-sky-12" } },
  ],
  defaultVariants: {
    size: 2,
    variant: "soft",
    color: "neutral",
    highContrast: false,
  },
});

export const calloutSkeleton = tv({
  base: "w-full animate-pulse rounded-3 bg-sand-3",
  variants: {
    size: {
      1: "h-9",
      2: "h-11",
      3: "h-14",
    },
  },
  defaultVariants: {
    size: 2,
  },
});
