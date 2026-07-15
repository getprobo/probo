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

// Toaster (Radix "Toast" over Base UI's Toast). A bottom-right stack of
// dismissible notifications, colored by type.

export const toaster = tv({
  slots: {
    viewport: "fixed right-4 bottom-4 z-50 flex w-full max-w-sm flex-col gap-2 outline-none",
    toast: [
      "pointer-events-auto flex items-start gap-3 rounded-3 border border-sand-a6 bg-sand-1 p-3 shadow-3",
      "transition-all duration-200 ease-out",
      "data-starting-style:translate-x-4 data-starting-style:opacity-0",
      "data-ending-style:translate-x-4 data-ending-style:opacity-0",
    ],
    content: "flex min-w-0 flex-1 flex-col gap-1",
    title: "text-2 font-medium text-sand-12",
    description: "text-2 break-words text-sand-11",
    close: "flex size-5 shrink-0 items-center justify-center rounded-1 text-sand-a10 outline-none transition-colors hover:bg-sand-3 hover:text-sand-12 [&_svg]:size-4",
  },
  variants: {
    type: {
      success: { toast: "border-green-6 bg-green-2", title: "text-green-12" },
      error: { toast: "border-red-6 bg-red-2", title: "text-red-12" },
      neutral: {},
    },
  },
  defaultVariants: {
    type: "neutral",
  },
});
