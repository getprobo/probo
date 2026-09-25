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

export const editableAvatarButton = tv({
  slots: {
    root: [
      "group relative inline-flex shrink-0 cursor-pointer outline-none",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
    frame: "relative overflow-hidden",
    overlay: [
      "pointer-events-none absolute inset-0 flex items-center justify-center",
      "bg-sand-12/50 text-sand-1 opacity-0 transition-opacity",
      "group-hover:opacity-100 group-focus-visible:opacity-100",
    ],
    badge: [
      "absolute -right-1 -bottom-1 z-1",
      "flex size-3.5 items-center justify-center rounded-full",
      "border border-sand-6 bg-sand-1 text-sand-12 shadow-1",
    ],
  },
  variants: {
    radius: {
      small: { root: "rounded-2", frame: "rounded-2" },
      full: { root: "rounded-full", frame: "rounded-full" },
    },
  },
  defaultVariants: {
    radius: "small",
  },
});
