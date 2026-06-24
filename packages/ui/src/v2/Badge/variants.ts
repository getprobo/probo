// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { tv } from "tailwind-variants/lite";

// Small status label (Radix "Badge"). A non-interactive <span>. The
// variant × color surface treatment resolves in the compound variants below.
export const badge = tv({
  base: "inline-flex w-fit items-center justify-center border border-transparent font-medium whitespace-nowrap align-middle",
  variants: {
    // Corner radius is bound to size (per the Figma, radius is global/theme,
    // not a per-instance prop).
    size: {
      1: "gap-1 rounded-1 px-1.5 py-0.5 text-1 [&_svg]:size-3",
      2: "gap-1 rounded-2 px-2 py-0.5 text-2 [&_svg]:size-4",
      3: "gap-1.5 rounded-2 px-2.5 py-1 text-3 [&_svg]:size-4",
    },
    variant: {
      solid: "",
      soft: "",
      surface: "",
      outline: "",
    },
    color: {
      neutral: "",
      gold: "",
      red: "",
      green: "",
      amber: "",
      sky: "",
    },
    highContrast: {
      true: "",
      false: "",
    },
  },
  compoundVariants: [
    // solid: filled step-9 background, white text (dark for light hues)
    { variant: "solid", color: "neutral", class: "bg-sand-9 text-white" },
    { variant: "solid", color: "gold", class: "bg-gold-9 text-white" },
    { variant: "solid", color: "red", class: "bg-red-9 text-white" },
    { variant: "solid", color: "green", class: "bg-green-9 text-white" },
    { variant: "solid", color: "amber", class: "bg-amber-9 text-amber-12" },
    { variant: "solid", color: "sky", class: "bg-sky-9 text-sky-12" },
    { variant: "solid", color: "neutral", highContrast: true, class: "bg-sand-12 text-sand-1" },
    { variant: "solid", color: "gold", highContrast: true, class: "bg-gold-12 text-gold-1" },
    { variant: "solid", color: "red", highContrast: true, class: "bg-red-12 text-red-1" },
    { variant: "solid", color: "green", highContrast: true, class: "bg-green-12 text-green-1" },
    { variant: "solid", color: "amber", highContrast: true, class: "bg-amber-12 text-amber-1" },
    { variant: "solid", color: "sky", highContrast: true, class: "bg-sky-12 text-sky-1" },

    // soft: tinted step-3 background
    { variant: "soft", color: "neutral", class: "bg-sand-3 text-sand-11" },
    { variant: "soft", color: "gold", class: "bg-gold-3 text-gold-11" },
    { variant: "soft", color: "red", class: "bg-red-3 text-red-11" },
    { variant: "soft", color: "green", class: "bg-green-3 text-green-11" },
    { variant: "soft", color: "amber", class: "bg-amber-3 text-amber-11" },
    { variant: "soft", color: "sky", class: "bg-sky-3 text-sky-11" },

    // surface: subtle background + border
    { variant: "surface", color: "neutral", class: "bg-sand-2 border-sand-6 text-sand-11" },
    { variant: "surface", color: "gold", class: "bg-gold-2 border-gold-6 text-gold-11" },
    { variant: "surface", color: "red", class: "bg-red-2 border-red-6 text-red-11" },
    { variant: "surface", color: "green", class: "bg-green-2 border-green-6 text-green-11" },
    { variant: "surface", color: "amber", class: "bg-amber-2 border-amber-6 text-amber-11" },
    { variant: "surface", color: "sky", class: "bg-sky-2 border-sky-6 text-sky-11" },

    // outline: border only
    { variant: "outline", color: "neutral", class: "border-sand-6 text-sand-11" },
    { variant: "outline", color: "gold", class: "border-gold-6 text-gold-11" },
    { variant: "outline", color: "red", class: "border-red-6 text-red-11" },
    { variant: "outline", color: "green", class: "border-green-6 text-green-11" },
    { variant: "outline", color: "amber", class: "border-amber-6 text-amber-11" },
    { variant: "outline", color: "sky", class: "border-sky-6 text-sky-11" },

    // high-contrast text for the tinted variants
    { variant: ["soft", "surface", "outline"], color: "neutral", highContrast: true, class: "text-sand-12" },
    { variant: ["soft", "surface", "outline"], color: "gold", highContrast: true, class: "text-gold-12" },
    { variant: ["soft", "surface", "outline"], color: "red", highContrast: true, class: "text-red-12" },
    { variant: ["soft", "surface", "outline"], color: "green", highContrast: true, class: "text-green-12" },
    { variant: ["soft", "surface", "outline"], color: "amber", highContrast: true, class: "text-amber-12" },
    { variant: ["soft", "surface", "outline"], color: "sky", highContrast: true, class: "text-sky-12" },
  ],
  defaultVariants: {
    size: 1,
    variant: "soft",
    color: "neutral",
    highContrast: false,
  },
});

export const badgeSkeleton = tv({
  base: "inline-block shrink-0 animate-pulse bg-sand-3 align-middle",
  variants: {
    size: {
      1: "h-5 w-12 rounded-1",
      2: "h-6 w-14 rounded-2",
      3: "h-7 w-16 rounded-2",
    },
  },
  defaultVariants: {
    size: 1,
  },
});
