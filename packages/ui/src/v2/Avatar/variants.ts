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

// Profile picture / initials / fallback icon (Radix "Avatar").
//   root     dimensions + radius + the fallback surface color
//   image    fills the root, cropped to cover
//   fallback initials/icon shown until the image loads
export const avatar = tv({
  slots: {
    root: "inline-flex shrink-0 items-center justify-center overflow-hidden align-middle select-none",
    image: "size-full object-cover",
    fallback: "flex size-full items-center justify-center font-medium leading-none",
  },
  variants: {
    size: {
      1: { root: "size-6", fallback: "text-1" },
      2: { root: "size-8", fallback: "text-2" },
      3: { root: "size-10", fallback: "text-3" },
      4: { root: "size-12", fallback: "text-3" },
      5: { root: "size-16", fallback: "text-5" },
      6: { root: "size-20", fallback: "text-6" },
      7: { root: "size-24", fallback: "text-7" },
      8: { root: "size-32", fallback: "text-8" },
      9: { root: "size-40", fallback: "text-9" },
    },
    radius: {
      none: { root: "rounded-none" },
      small: { root: "rounded-2" },
      medium: { root: "rounded-3" },
      large: { root: "rounded-4" },
      full: { root: "rounded-full" },
    },
    // Surface treatment + hue resolve together in the compound variants below.
    variant: {
      solid: {},
      soft: {},
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
    // Soft: tinted background (step 3), hue text (step 11 / 12 high-contrast).
    { variant: "soft", color: "neutral", highContrast: false, class: { fallback: "bg-sand-3 text-sand-11" } },
    { variant: "soft", color: "neutral", highContrast: true, class: { fallback: "bg-sand-3 text-sand-12" } },
    { variant: "soft", color: "gold", highContrast: false, class: { fallback: "bg-gold-3 text-gold-11" } },
    { variant: "soft", color: "gold", highContrast: true, class: { fallback: "bg-gold-3 text-gold-12" } },
    { variant: "soft", color: "red", highContrast: false, class: { fallback: "bg-red-3 text-red-11" } },
    { variant: "soft", color: "red", highContrast: true, class: { fallback: "bg-red-3 text-red-12" } },
    { variant: "soft", color: "green", highContrast: false, class: { fallback: "bg-green-3 text-green-11" } },
    { variant: "soft", color: "green", highContrast: true, class: { fallback: "bg-green-3 text-green-12" } },
    { variant: "soft", color: "amber", highContrast: false, class: { fallback: "bg-amber-3 text-amber-11" } },
    { variant: "soft", color: "amber", highContrast: true, class: { fallback: "bg-amber-3 text-amber-12" } },
    { variant: "soft", color: "sky", highContrast: false, class: { fallback: "bg-sky-3 text-sky-11" } },
    { variant: "soft", color: "sky", highContrast: true, class: { fallback: "bg-sky-3 text-sky-12" } },

    // Solid: filled background (step 9, or 10 high-contrast). Most hues take
    // white text; amber/sky steps 9-10 are light and take dark text.
    { variant: "solid", color: "neutral", highContrast: false, class: { fallback: "bg-sand-9 text-white" } },
    { variant: "solid", color: "neutral", highContrast: true, class: { fallback: "bg-sand-10 text-white" } },
    { variant: "solid", color: "gold", highContrast: false, class: { fallback: "bg-gold-9 text-white" } },
    { variant: "solid", color: "gold", highContrast: true, class: { fallback: "bg-gold-10 text-white" } },
    { variant: "solid", color: "red", highContrast: false, class: { fallback: "bg-red-9 text-white" } },
    { variant: "solid", color: "red", highContrast: true, class: { fallback: "bg-red-10 text-white" } },
    { variant: "solid", color: "green", highContrast: false, class: { fallback: "bg-green-9 text-white" } },
    { variant: "solid", color: "green", highContrast: true, class: { fallback: "bg-green-10 text-white" } },
    { variant: "solid", color: "amber", highContrast: false, class: { fallback: "bg-amber-9 text-amber-12" } },
    { variant: "solid", color: "amber", highContrast: true, class: { fallback: "bg-amber-10 text-amber-12" } },
    { variant: "solid", color: "sky", highContrast: false, class: { fallback: "bg-sky-9 text-sky-12" } },
    { variant: "solid", color: "sky", highContrast: true, class: { fallback: "bg-sky-10 text-sky-12" } },
  ],
  defaultVariants: {
    size: 3,
    variant: "soft",
    color: "neutral",
    highContrast: false,
    radius: "medium",
  },
});

export const avatarSkeleton = tv({
  base: "inline-block shrink-0 animate-pulse bg-sand-3 align-middle",
  variants: {
    size: {
      1: "size-6",
      2: "size-8",
      3: "size-10",
      4: "size-12",
      5: "size-16",
      6: "size-20",
      7: "size-24",
      8: "size-32",
      9: "size-40",
    },
    radius: {
      none: "rounded-none",
      small: "rounded-2",
      medium: "rounded-3",
      large: "rounded-4",
      full: "rounded-full",
    },
  },
  defaultVariants: {
    size: 3,
    radius: "medium",
  },
});
