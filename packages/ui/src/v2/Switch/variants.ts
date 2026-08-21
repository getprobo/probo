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

// Toggle track (Radix "Switch"). Checked / hover / active / disabled resolve
// off Base UI's data-* attributes. Figma ships accent (gold) only; extra hues
// are the same compounds with the hue swapped (per the Switch docs page).
export const switchRoot = tv({
  base: [
    "inline-flex shrink-0 cursor-pointer items-center rounded-full border p-px",
    "outline-none transition-colors select-none",
    "focus-visible:ring-2",
    "data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50",
  ],
  variants: {
    size: {
      1: "h-5 w-[28px]",
      2: "h-5 w-[35px]",
      3: "h-6 w-[42px]",
    },
    variant: {
      classic: "",
      surface: "",
      soft: "",
    },
    color: {
      gold: "focus-visible:ring-gold-8",
      green: "focus-visible:ring-green-8",
    },
    highContrast: {
      true: "",
      false: "",
    },
  },
  compoundVariants: [
    // surface (default): sand-alpha track, hue-9 when on
    {
      variant: "surface",
      color: "gold",
      highContrast: false,
      class: [
        "border-sand-a7 bg-sand-a3",
        "hover:bg-sand-a4 active:bg-sand-a5",
        "data-[checked]:border-transparent data-[checked]:bg-gold-9",
        "data-[checked]:hover:bg-gold-10 data-[checked]:active:bg-gold-10",
      ],
    },
    {
      variant: "surface",
      color: "green",
      highContrast: false,
      class: [
        "border-sand-a7 bg-sand-a3",
        "hover:bg-sand-a4 active:bg-sand-a5",
        "data-[checked]:border-transparent data-[checked]:bg-green-9",
        "data-[checked]:hover:bg-green-10 data-[checked]:active:bg-green-10",
      ],
    },
    {
      variant: "surface",
      color: "gold",
      highContrast: true,
      class: [
        "border-sand-a7 bg-sand-a3",
        "hover:bg-sand-a4 active:bg-sand-a5",
        "data-[checked]:border-transparent data-[checked]:bg-gold-12",
      ],
    },
    {
      variant: "surface",
      color: "green",
      highContrast: true,
      class: [
        "border-sand-a7 bg-sand-a3",
        "hover:bg-sand-a4 active:bg-sand-a5",
        "data-[checked]:border-transparent data-[checked]:bg-green-12",
      ],
    },

    // classic: same surfaces plus a recessed inset
    {
      variant: "classic",
      color: "gold",
      highContrast: false,
      class: [
        "border-sand-a6 bg-sand-a3 inset-shadow-1",
        "hover:bg-sand-a4 active:bg-sand-a5",
        "data-[checked]:border-sand-a7 data-[checked]:bg-gold-9",
        "data-[checked]:hover:bg-gold-10 data-[checked]:active:bg-gold-10",
      ],
    },
    {
      variant: "classic",
      color: "green",
      highContrast: false,
      class: [
        "border-sand-a6 bg-sand-a3 inset-shadow-1",
        "hover:bg-sand-a4 active:bg-sand-a5",
        "data-[checked]:border-sand-a7 data-[checked]:bg-green-9",
        "data-[checked]:hover:bg-green-10 data-[checked]:active:bg-green-10",
      ],
    },
    {
      variant: "classic",
      color: "gold",
      highContrast: true,
      class: [
        "border-sand-a6 bg-sand-a3 inset-shadow-1",
        "hover:bg-sand-a4 active:bg-sand-a5",
        "data-[checked]:border-sand-a7 data-[checked]:bg-gold-12",
      ],
    },
    {
      variant: "classic",
      color: "green",
      highContrast: true,
      class: [
        "border-sand-a6 bg-sand-a3 inset-shadow-1",
        "hover:bg-sand-a4 active:bg-sand-a5",
        "data-[checked]:border-sand-a7 data-[checked]:bg-green-12",
      ],
    },

    // soft: tinted hue, no border
    {
      variant: "soft",
      color: "gold",
      highContrast: false,
      class: [
        "border-transparent bg-gold-4",
        "hover:bg-gold-5 active:bg-gold-5",
        "data-[checked]:bg-gold-8",
        "data-[checked]:hover:bg-gold-9 data-[checked]:active:bg-gold-9",
      ],
    },
    {
      variant: "soft",
      color: "green",
      highContrast: false,
      class: [
        "border-transparent bg-green-4",
        "hover:bg-green-5 active:bg-green-5",
        "data-[checked]:bg-green-8",
        "data-[checked]:hover:bg-green-9 data-[checked]:active:bg-green-9",
      ],
    },
    {
      variant: "soft",
      color: "gold",
      highContrast: true,
      class: [
        "border-transparent bg-gold-4",
        "hover:bg-gold-5 active:bg-gold-5",
        "data-[checked]:bg-gold-11",
      ],
    },
    {
      variant: "soft",
      color: "green",
      highContrast: true,
      class: [
        "border-transparent bg-green-4",
        "hover:bg-green-5 active:bg-green-5",
        "data-[checked]:bg-green-11",
      ],
    },
  ],
  defaultVariants: {
    size: 2,
    variant: "surface",
    color: "gold",
    highContrast: false,
  },
});

export const switchThumb = tv({
  base: [
    "rounded-full bg-white shadow-2",
    "border border-sand-8",
    "transition-transform duration-150 ease-out",
    "motion-reduce:transition-none",
    "data-[checked]:border-black/20",
  ],
  variants: {
    // Travel = track width − vertical padding (2×1px) − thumb. justify-end
    // cannot interpolate; translate is what makes the thumb slide.
    size: {
      1: "size-[18px] data-[checked]:translate-x-[8px] rtl:data-[checked]:-translate-x-[8px]",
      2: "size-[18px] data-[checked]:translate-x-[15px] rtl:data-[checked]:-translate-x-[15px]",
      3: "size-[22px] data-[checked]:translate-x-[18px] rtl:data-[checked]:-translate-x-[18px]",
    },
  },
  defaultVariants: {
    size: 2,
  },
});

export const switchSkeleton = tv({
  base: "inline-block shrink-0 animate-pulse rounded-full bg-sand-3 align-middle",
  variants: {
    size: {
      1: "h-5 w-[28px]",
      2: "h-5 w-[35px]",
      3: "h-6 w-[42px]",
    },
  },
  defaultVariants: {
    size: 2,
  },
});
