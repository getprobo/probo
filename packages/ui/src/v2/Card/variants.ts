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

// Container that groups related content (Radix "Card"). The surface treatment
// resolves in the compound variants below.
export const card = tv({
  base: "block overflow-hidden",
  variants: {
    // Radius scales with size (per the Figma).
    size: {
      1: "rounded-4",
      2: "rounded-4",
      3: "rounded-5",
      4: "rounded-5",
      5: "rounded-6",
    },
    // Padding defaults to match `size` (set in Card.tsx); pass `padding="none"`
    // for cards whose regions own their own padding (e.g. full-bleed media).
    padding: {
      1: "p-3",
      2: "p-4",
      3: "p-5",
      4: "p-6",
      5: "p-8",
      none: "p-0",
    },
    variant: {
      surface: "border border-sand-6 bg-sand-2",
      classic: "border border-sand-6 bg-sand-2 shadow-2",
      // Soft surface: translucent neutral border on the app background.
      soft: "border border-sand-a3 bg-sand-1",
      ghost: "",
    },
    // Hover/active affordance for clickable cards. This tunes look only; wrap
    // the card in an <a>/router link (or Base UI `render`) for real navigation.
    interactive: {
      true: "cursor-pointer transition-colors",
      false: "",
    },
  },
  compoundVariants: [
    { variant: ["surface", "classic", "soft"], interactive: true, class: "hover:bg-sand-3 active:bg-sand-4" },
    { variant: "ghost", interactive: true, class: "hover:bg-sand-3 active:bg-sand-4" },
  ],
  defaultVariants: {
    size: 1,
    variant: "surface",
    interactive: false,
  },
});

// Negates the card's padding so content bleeds to its edges. The negative
// margin must match the card's resolved padding (1→3, 2→4, 3→5, 4→6, 5→8), so
// the matrix is keyed on the padding (from context) and side. `padding="none"`
// has nothing to negate. `overflow-hidden` clips bled media.
export const cardInset = tv({
  base: "overflow-hidden",
  variants: {
    padding: { 1: "", 2: "", 3: "", 4: "", 5: "", none: "" },
    side: { all: "", x: "", y: "", top: "", bottom: "" },
  },
  compoundVariants: [
    { padding: 1, side: "all", class: "-m-3" },
    { padding: 1, side: "x", class: "-mx-3" },
    { padding: 1, side: "y", class: "-my-3" },
    { padding: 1, side: "top", class: "-mx-3 -mt-3 mb-3" },
    { padding: 1, side: "bottom", class: "-mx-3 -mb-3 mt-3" },
    { padding: 2, side: "all", class: "-m-4" },
    { padding: 2, side: "x", class: "-mx-4" },
    { padding: 2, side: "y", class: "-my-4" },
    { padding: 2, side: "top", class: "-mx-4 -mt-4 mb-4" },
    { padding: 2, side: "bottom", class: "-mx-4 -mb-4 mt-4" },
    { padding: 3, side: "all", class: "-m-5" },
    { padding: 3, side: "x", class: "-mx-5" },
    { padding: 3, side: "y", class: "-my-5" },
    { padding: 3, side: "top", class: "-mx-5 -mt-5 mb-5" },
    { padding: 3, side: "bottom", class: "-mx-5 -mb-5 mt-5" },
    { padding: 4, side: "all", class: "-m-6" },
    { padding: 4, side: "x", class: "-mx-6" },
    { padding: 4, side: "y", class: "-my-6" },
    { padding: 4, side: "top", class: "-mx-6 -mt-6 mb-6" },
    { padding: 4, side: "bottom", class: "-mx-6 -mb-6 mt-6" },
    { padding: 5, side: "all", class: "-m-8" },
    { padding: 5, side: "x", class: "-mx-8" },
    { padding: 5, side: "y", class: "-my-8" },
    { padding: 5, side: "top", class: "-mx-8 -mt-8 mb-8" },
    { padding: 5, side: "bottom", class: "-mx-8 -mb-8 mt-8" },
  ],
  defaultVariants: {
    padding: 1,
    side: "all",
  },
});

export const cardSkeleton = tv({
  base: "block w-full animate-pulse rounded-4 bg-sand-3",
  variants: {
    size: {
      1: "h-20",
      2: "h-24",
      3: "h-28",
      4: "h-32",
      5: "h-40",
    },
  },
  defaultVariants: {
    size: 1,
  },
});
