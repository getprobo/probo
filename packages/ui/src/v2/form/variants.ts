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

// Text input (Radix "Text Field" over Base UI's Input). A bordered surface on
// the neutral background with an optional leading icon slot.
export const textField = tv({
  slots: {
    root: [
      "flex items-center gap-2 rounded-2 text-2 text-sand-12 transition-colors",
      "focus-within:ring-2 focus-within:ring-sand-8 focus-within:ring-offset-1 focus-within:ring-offset-sand-1",
      "has-[input:disabled]:pointer-events-none has-[input:disabled]:opacity-50",
    ],
    icon: "flex size-4 shrink-0 items-center justify-center text-sand-a9 [&_svg]:size-4",
    input: "min-w-0 flex-1 bg-transparent text-sand-12 outline-none placeholder:text-sand-a9",
  },
  variants: {
    size: {
      1: { root: "h-7 px-2" },
      2: { root: "h-8 px-2" },
    },
    // Surface treatment. Only the accent (gold) color ships, matching Figma;
    // classic adds a recessed inset shadow, soft drops the border for a tint.
    variant: {
      classic: { root: "border border-sand-a5 bg-sand-1 inset-shadow-2" },
      surface: { root: "border border-sand-a5 bg-sand-1" },
      soft: { root: "bg-gold-3" },
    },
  },
  defaultVariants: {
    size: 2,
    variant: "surface",
  },
});

// Multi-line text input, mirroring TextField's bordered surface. Base UI has no
// textarea primitive, so this styles a native <textarea>.
export const textArea = tv({
  slots: {
    root: [
      "flex rounded-2 text-2 text-sand-12 transition-colors",
      "focus-within:ring-2 focus-within:ring-sand-8 focus-within:ring-offset-1 focus-within:ring-offset-sand-1",
      "has-[textarea:disabled]:pointer-events-none has-[textarea:disabled]:opacity-50",
    ],
    textarea: [
      "min-h-16 w-full resize-y bg-transparent px-2 py-1.5 text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
  },
  variants: {
    variant: {
      classic: { root: "border border-sand-a5 bg-sand-1 inset-shadow-2" },
      surface: { root: "border border-sand-a5 bg-sand-1" },
      soft: { root: "bg-gold-3" },
    },
  },
  defaultVariants: {
    variant: "surface",
  },
});

// Vertical label + control + error grouping used by form dialogs.
export const field = tv({
  slots: {
    root: "flex flex-col gap-1.5",
    label: "flex flex-col gap-1.5",
    labelText: "text-2 font-medium text-sand-12",
    error: "text-1 text-red-a11",
  },
});

export const textFieldSkeleton = tv({
  base: "inline-block animate-pulse rounded-2 bg-sand-3 align-middle",
  variants: {
    size: {
      1: "h-7 w-60",
      2: "h-8 w-60",
    },
  },
  defaultVariants: {
    size: 2,
  },
});
