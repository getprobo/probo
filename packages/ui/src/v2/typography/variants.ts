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

// Numbered type scale (text-1 … text-9). Each utility carries its paired
// font-size, line-height, and letter-spacing from the v2 theme. Shared across
// the typography primitives and their skeletons so placeholders match the real
// line height.
const size = {
  1: "text-1",
  2: "text-2",
  3: "text-3",
  4: "text-4",
  5: "text-5",
  6: "text-6",
  7: "text-7",
  8: "text-8",
  9: "text-9",
} as const;

// Shared look across Text and Heading: only their defaults differ.
const typographyVariants = {
  size,
  weight: {
    light: "font-light",
    regular: "font-normal",
    medium: "font-medium",
    bold: "font-bold",
  },
  align: {
    left: "text-left",
    center: "text-center",
    right: "text-right",
  },
  // Hue only; the resolved text step comes from the color × highContrast
  // compound variants below (step 11 low-contrast, step 12 high-contrast).
  // `faint` is a de-emphasized neutral (sand alpha 8) for fine-print metadata.
  color: {
    neutral: "",
    faint: "",
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
};

const colorCompoundVariants = [
  { color: "neutral", highContrast: false, class: "text-sand-11" },
  { color: "neutral", highContrast: true, class: "text-sand-12" },
  // Faint metadata; a single de-emphasized step regardless of highContrast.
  { color: "faint", highContrast: false, class: "text-sand-a8" },
  { color: "faint", highContrast: true, class: "text-sand-a8" },
  { color: "gold", highContrast: false, class: "text-gold-11" },
  { color: "gold", highContrast: true, class: "text-gold-12" },
  { color: "red", highContrast: false, class: "text-red-11" },
  { color: "red", highContrast: true, class: "text-red-12" },
  { color: "green", highContrast: false, class: "text-green-11" },
  { color: "green", highContrast: true, class: "text-green-12" },
  { color: "amber", highContrast: false, class: "text-amber-11" },
  { color: "amber", highContrast: true, class: "text-amber-12" },
  { color: "sky", highContrast: false, class: "text-sky-11" },
  { color: "sky", highContrast: true, class: "text-sky-12" },
] as const;

export const text = tv({
  variants: typographyVariants,
  compoundVariants: [...colorCompoundVariants],
  defaultVariants: {
    size: 3,
    weight: "regular",
    color: "neutral",
    highContrast: false,
  },
});

export const heading = tv({
  variants: typographyVariants,
  compoundVariants: [...colorCompoundVariants],
  defaultVariants: {
    size: 6,
    weight: "bold",
    color: "neutral",
    highContrast: false,
  },
});

const skeletonBase = "inline-block animate-pulse select-none rounded-2 bg-sand-3 text-transparent";

export const textSkeleton = tv({
  base: skeletonBase,
  variants: {
    size,
  },
  defaultVariants: {
    size: 3,
  },
});

export const headingSkeleton = tv({
  base: skeletonBase,
  variants: {
    size,
  },
  defaultVariants: {
    size: 6,
  },
});
