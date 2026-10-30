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
// the neutral background with an optional leading icon and trailing suffix.
export const textField = tv({
  slots: {
    root: [
      "flex items-center rounded-2 text-sand-12 transition-colors",
      "focus-within:ring-2 focus-within:ring-sand-8 focus-within:ring-offset-1 focus-within:ring-offset-sand-1",
      "has-[input:disabled]:pointer-events-none has-[input:disabled]:opacity-50",
    ],
    icon: "flex shrink-0 items-center justify-center text-sand-a9",
    input: "min-w-0 flex-1 bg-transparent text-sand-12 outline-none placeholder:text-sand-a9",
    suffix: "pointer-events-none flex shrink-0 items-center whitespace-nowrap text-sand-a11 select-none",
  },
  variants: {
    size: {
      1: {
        root: "h-6 gap-1 px-1.5 text-1",
        icon: "size-3.5 [&_svg]:size-3.5",
      },
      2: {
        root: "h-8 gap-2 px-2 text-2",
        icon: "size-4 [&_svg]:size-4",
      },
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
    labelText: "text-2 font-medium text-sand-12",
    requiredMark: "ms-0.5 text-red-a11",
    error: "text-1 text-red-a11",
  },
});

export const textFieldSkeleton = tv({
  base: "inline-block animate-pulse rounded-2 bg-sand-3 align-middle",
  variants: {
    size: {
      1: "h-6 w-60",
      2: "h-8 w-60",
    },
  },
  defaultVariants: {
    size: 2,
  },
});

// Date picker trigger (TextField-like surface + typed input) plus the calendar popup chrome.
export const dateField = tv({
  slots: {
    root: "w-full",
    surface: [
      "flex min-w-0 w-full items-center gap-2 rounded-2 text-2 text-sand-12 transition-colors",
      "focus-within:ring-2 focus-within:ring-sand-8 focus-within:ring-offset-1 focus-within:ring-offset-sand-1",
      "has-data-popup-open:ring-2 has-data-popup-open:ring-sand-8 has-data-popup-open:ring-offset-1 has-data-popup-open:ring-offset-sand-1",
      "has-[input:disabled]:pointer-events-none has-[input:disabled]:opacity-50",
    ],
    glyph: "flex size-4 shrink-0 items-center justify-center text-sand-a9 [&_svg]:size-4",
    rangeButton: [
      "flex h-full min-w-0 flex-1 cursor-pointer items-center gap-2 border-0 bg-transparent p-0",
      "text-start outline-none",
      "disabled:pointer-events-none",
    ],
    caption: "min-w-0 flex-1 truncate text-start",
    popup: "flex flex-col gap-2",
    iconTrigger: [
      "flex size-4 shrink-0 items-center justify-center border-0 bg-transparent p-0",
      "cursor-pointer text-sand-a9 outline-none [&_svg]:size-4 hover:text-sand-12",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
      "data-disabled:pointer-events-none data-disabled:opacity-50",
    ],
    input: [
      "min-w-0 flex-1 bg-transparent text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
    clear: [
      "flex size-6 shrink-0 items-center justify-center rounded-2 border-0 bg-transparent p-0",
      "cursor-pointer text-sand-a9 outline-none [&_svg]:size-4 hover:text-sand-12 hover:bg-sand-3",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
    calendar: "flex flex-col",
    header: "flex items-center gap-1",
    monthSelect: "min-w-0",
    yearSelect: "shrink-0",
    weekdays: "grid grid-cols-7",
    weekday: "py-1 text-center text-1 font-medium text-sand-11",
    grid: "grid grid-cols-7",
  },
  variants: {
    size: {
      1: {
        surface: "h-7 px-2",
        calendar: "w-56 gap-1",
        monthSelect: "w-20 shrink-0",
        yearSelect: "w-16",
      },
      2: {
        surface: "h-8 px-2",
        calendar: "w-72 gap-2",
        monthSelect: "flex-1",
        yearSelect: "w-20",
      },
    },
    variant: {
      classic: { surface: "border border-sand-a5 bg-sand-1 inset-shadow-2" },
      surface: { surface: "border border-sand-a5 bg-sand-1" },
    },
    filled: {
      true: { caption: "text-sand-12" },
      false: { caption: "text-sand-a9" },
    },
  },
  defaultVariants: {
    size: 2,
    variant: "surface",
    filled: true,
  },
});

export const dateFieldDay = tv({
  base: [
    "flex items-center justify-center text-sand-12",
    "cursor-pointer outline-none",
    "focus-visible:ring-2 focus-visible:ring-sand-8",
    "disabled:pointer-events-none disabled:opacity-50",
  ],
  variants: {
    size: {
      1: "size-6 text-1",
      2: "size-8 text-2",
    },
    // One background and one radius per tone. lite has no class merge, so a
    // day must not emit both hover:bg-sand-4 and a selected background.
    tone: {
      day: "rounded-2 hover:bg-sand-4",
      selected: "rounded-2 bg-sand-5 hover:bg-sand-5",
      within: "rounded-none bg-sand-3 hover:bg-sand-4",
      start: "rounded-s-2 bg-sand-12 text-sand-1 hover:bg-sand-12",
      end: "rounded-e-2 bg-sand-12 text-sand-1 hover:bg-sand-12",
      single: "rounded-2 bg-sand-12 text-sand-1 hover:bg-sand-12",
    },
    today: {
      true: "font-medium",
      false: "",
    },
    outside: {
      true: "text-sand-a9",
      false: "",
    },
  },
  defaultVariants: {
    size: 2,
    tone: "day",
    today: false,
    outside: false,
  },
});

export const dateFieldSkeleton = tv({
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

// Always-visible calendar panel: typed input above DateMonthView.
export const dateCalendar = tv({
  slots: {
    root: "flex min-w-0 flex-col",
    surface: [
      "flex min-w-0 w-full items-center gap-2 rounded-2 text-sand-12 transition-colors",
      "focus-within:ring-2 focus-within:ring-sand-8 focus-within:ring-offset-1 focus-within:ring-offset-sand-1",
      "has-[input:disabled]:pointer-events-none has-[input:disabled]:opacity-50",
    ],
    input: [
      "min-w-0 w-full flex-1 bg-transparent text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
    clear: [
      "flex shrink-0 items-center justify-center rounded-2 border-0 bg-transparent p-0",
      "cursor-pointer text-sand-a9 outline-none hover:text-sand-12 hover:bg-sand-3",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
  },
  variants: {
    size: {
      1: {
        root: "w-56 gap-1",
        surface: "h-7 px-2 text-1",
        clear: "size-5 [&_svg]:size-3",
      },
      2: {
        root: "w-72 gap-2",
        surface: "h-8 px-2 text-2",
        clear: "size-6 [&_svg]:size-4",
      },
    },
    variant: {
      classic: { surface: "border border-sand-a5 bg-sand-1 inset-shadow-2" },
      surface: { surface: "border border-sand-a5 bg-sand-1" },
    },
  },
  defaultVariants: {
    size: 2,
    variant: "surface",
  },
});
