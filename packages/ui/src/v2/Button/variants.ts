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

// Button (Radix "Button"). A <button>; an Anchor / Link are separate
// components (see contrib/claude/ui.md). The variant × color surface treatment
// resolves in the compound variants below.
export const button = tv({
  base: [
    "inline-flex shrink-0 items-center justify-center border border-transparent font-medium whitespace-nowrap",
    "cursor-pointer outline-none transition-colors select-none",
    "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    "disabled:pointer-events-none disabled:opacity-50",
  ],
  variants: {
    // Corner radius is bound to size (per the Figma, radius is global/theme,
    // not a per-instance prop).
    size: {
      1: "h-6 gap-1 rounded-2 px-2 text-1 [&_svg]:size-4",
      2: "h-8 gap-2 rounded-2 px-3 text-2 [&_svg]:size-4",
      3: "h-10 gap-2 rounded-3 px-4 text-3 [&_svg]:size-5",
      4: "h-12 gap-2 rounded-3 px-5 text-4 [&_svg]:size-5",
    },
    variant: {
      classic: "",
      solid: "",
      soft: "",
      surface: "",
      outline: "",
      ghost: "",
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
    // Persistent "selected" treatment for nav items built on Anchor / Link
    // (ghost / soft surfaces). Look-only — does not change structure.
    active: {
      true: "",
      false: "",
    },
  },
  compoundVariants: [
    // ── solid: filled step-9 background, white text (dark for light hues) ──
    // Gated on highContrast:false so the high-contrast rules below don't also
    // apply (tailwind-variants/lite has no merge — both would emit and a static
    // text-white would win, breaking dark mode).
    { variant: ["solid", "classic"], color: "neutral", highContrast: false, class: "bg-sand-9 text-white hover:bg-sand-10" },
    { variant: ["solid", "classic"], color: "gold", highContrast: false, class: "bg-gold-9 text-white hover:bg-gold-10" },
    { variant: ["solid", "classic"], color: "red", highContrast: false, class: "bg-red-9 text-white hover:bg-red-10" },
    { variant: ["solid", "classic"], color: "green", highContrast: false, class: "bg-green-9 text-white hover:bg-green-10" },
    { variant: ["solid", "classic"], color: "amber", highContrast: false, class: "bg-amber-9 text-amber-12 hover:bg-amber-10" },
    { variant: ["solid", "classic"], color: "sky", highContrast: false, class: "bg-sky-9 text-sky-12 hover:bg-sky-10" },
    // classic adds elevation over solid
    { variant: "classic", class: "shadow-2" },
    // solid high-contrast: step-12 background, step-1 text
    { variant: ["solid", "classic"], color: "neutral", highContrast: true, class: "bg-sand-12 text-sand-1 hover:bg-sand-12" },
    { variant: ["solid", "classic"], color: "gold", highContrast: true, class: "bg-gold-12 text-gold-1 hover:bg-gold-12" },
    { variant: ["solid", "classic"], color: "red", highContrast: true, class: "bg-red-12 text-red-1 hover:bg-red-12" },
    { variant: ["solid", "classic"], color: "green", highContrast: true, class: "bg-green-12 text-green-1 hover:bg-green-12" },
    { variant: ["solid", "classic"], color: "amber", highContrast: true, class: "bg-amber-12 text-amber-1 hover:bg-amber-12" },
    { variant: ["solid", "classic"], color: "sky", highContrast: true, class: "bg-sky-12 text-sky-1 hover:bg-sky-12" },

    // ── soft: tinted step-3 background, step-11 text ──
    { variant: "soft", color: "neutral", class: "bg-sand-3 text-sand-11 hover:bg-sand-4" },
    { variant: "soft", color: "gold", class: "bg-gold-3 text-gold-11 hover:bg-gold-4" },
    { variant: "soft", color: "red", class: "bg-red-3 text-red-11 hover:bg-red-4" },
    { variant: "soft", color: "green", class: "bg-green-3 text-green-11 hover:bg-green-4" },
    { variant: "soft", color: "amber", class: "bg-amber-3 text-amber-11 hover:bg-amber-4" },
    { variant: "soft", color: "sky", class: "bg-sky-3 text-sky-11 hover:bg-sky-4" },

    // ── surface: subtle background + border ──
    { variant: "surface", color: "neutral", class: "bg-sand-2 text-sand-11 border-sand-6 hover:bg-sand-3" },
    { variant: "surface", color: "gold", class: "bg-gold-2 text-gold-11 border-gold-6 hover:bg-gold-3" },
    { variant: "surface", color: "red", class: "bg-red-2 text-red-11 border-red-6 hover:bg-red-3" },
    { variant: "surface", color: "green", class: "bg-green-2 text-green-11 border-green-6 hover:bg-green-3" },
    { variant: "surface", color: "amber", class: "bg-amber-2 text-amber-11 border-amber-6 hover:bg-amber-3" },
    { variant: "surface", color: "sky", class: "bg-sky-2 text-sky-11 border-sky-6 hover:bg-sky-3" },

    // ── outline: transparent background, step-7 border ──
    { variant: "outline", color: "neutral", class: "text-sand-11 border-sand-7 hover:bg-sand-3" },
    { variant: "outline", color: "gold", class: "text-gold-11 border-gold-7 hover:bg-gold-3" },
    { variant: "outline", color: "red", class: "text-red-11 border-red-7 hover:bg-red-3" },
    { variant: "outline", color: "green", class: "text-green-11 border-green-7 hover:bg-green-3" },
    { variant: "outline", color: "amber", class: "text-amber-11 border-amber-7 hover:bg-amber-3" },
    { variant: "outline", color: "sky", class: "text-sky-11 border-sky-7 hover:bg-sky-3" },

    // ── ghost: transparent until hover ──
    { variant: "ghost", color: "neutral", class: "text-sand-11 hover:bg-sand-3" },
    { variant: "ghost", color: "gold", class: "text-gold-11 hover:bg-gold-3" },
    { variant: "ghost", color: "red", class: "text-red-11 hover:bg-red-3" },
    { variant: "ghost", color: "green", class: "text-green-11 hover:bg-green-3" },
    { variant: "ghost", color: "amber", class: "text-amber-11 hover:bg-amber-3" },
    { variant: "ghost", color: "sky", class: "text-sky-11 hover:bg-sky-3" },

    // ── active: persistent selected background + high-contrast text ──
    // ghost goes from transparent to the step-3 tint; soft bumps to step-4.
    { variant: "ghost", color: "neutral", active: true, class: "bg-sand-3 text-sand-12" },
    { variant: "ghost", color: "gold", active: true, class: "bg-gold-3 text-gold-12" },
    { variant: "ghost", color: "red", active: true, class: "bg-red-3 text-red-12" },
    { variant: "ghost", color: "green", active: true, class: "bg-green-3 text-green-12" },
    { variant: "ghost", color: "amber", active: true, class: "bg-amber-3 text-amber-12" },
    { variant: "ghost", color: "sky", active: true, class: "bg-sky-3 text-sky-12" },
    { variant: "soft", color: "neutral", active: true, class: "bg-sand-4 text-sand-12" },
    { variant: "soft", color: "gold", active: true, class: "bg-gold-4 text-gold-12" },
    { variant: "soft", color: "red", active: true, class: "bg-red-4 text-red-12" },
    { variant: "soft", color: "green", active: true, class: "bg-green-4 text-green-12" },
    { variant: "soft", color: "amber", active: true, class: "bg-amber-4 text-amber-12" },
    { variant: "soft", color: "sky", active: true, class: "bg-sky-4 text-sky-12" },

    // ── high-contrast text bump for the tinted variants ──
    { variant: ["soft", "surface", "outline", "ghost"], color: "neutral", highContrast: true, class: "text-sand-12" },
    { variant: ["soft", "surface", "outline", "ghost"], color: "gold", highContrast: true, class: "text-gold-12" },
    { variant: ["soft", "surface", "outline", "ghost"], color: "red", highContrast: true, class: "text-red-12" },
    { variant: ["soft", "surface", "outline", "ghost"], color: "green", highContrast: true, class: "text-green-12" },
    { variant: ["soft", "surface", "outline", "ghost"], color: "amber", highContrast: true, class: "text-amber-12" },
    { variant: ["soft", "surface", "outline", "ghost"], color: "sky", highContrast: true, class: "text-sky-12" },
  ],
  defaultVariants: {
    size: 2,
    variant: "solid",
    color: "gold",
    highContrast: false,
    active: false,
  },
});

export const buttonSkeleton = tv({
  base: "inline-block shrink-0 animate-pulse bg-sand-3 align-middle",
  variants: {
    size: {
      1: "h-6 w-20 rounded-2",
      2: "h-8 w-24 rounded-2",
      3: "h-10 w-28 rounded-3",
      4: "h-12 w-32 rounded-3",
    },
  },
  defaultVariants: {
    size: 2,
  },
});
