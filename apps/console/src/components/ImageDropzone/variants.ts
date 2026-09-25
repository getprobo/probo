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

export const imageDropzone = tv({
  slots: {
    root: [
      "group relative flex items-center justify-center overflow-hidden rounded-4 border border-dashed",
      "border-sand-6 bg-sand-2 outline-none transition-colors",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1",
    ],
    image: "pointer-events-none absolute inset-0 size-full object-contain p-2",
    placeholder: "flex flex-col items-center gap-1 px-3 text-center",
    overlay: "pointer-events-none absolute inset-0 z-2 flex items-center justify-center bg-sand-1/70",
    hint: [
      "pointer-events-none absolute inset-0 z-1 hidden flex-col items-center justify-center gap-1 px-3 text-center",
      "bg-sand-1/70 backdrop-blur-sm group-hover:flex group-focus-visible:flex",
    ],
    clear: "absolute top-1 right-1 z-2",
  },
  variants: {
    ratio: {
      square: { root: "size-32" },
      wide: { root: "h-32 min-w-0 w-full" },
    },
    filled: {
      true: { root: "border-solid bg-sand-3" },
      false: {},
    },
    dragActive: {
      true: { root: "border-solid border-sky-8 bg-sky-2" },
      false: {},
    },
    disabled: {
      true: { root: "cursor-default opacity-60" },
      false: { root: "cursor-pointer hover:bg-sand-3" },
    },
  },
  defaultVariants: {
    ratio: "square",
    filled: false,
    dragActive: false,
    disabled: false,
  },
});
