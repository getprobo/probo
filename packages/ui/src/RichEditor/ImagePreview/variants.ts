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

// Auto margins center the popup without a transform, so the close control can
// stay position:fixed to the viewport and still live inside the focus trap.
export const imagePreview = tv({
  slots: {
    trigger: [
      "m-0 block w-full max-w-full cursor-zoom-in border-0 bg-transparent p-0",
      "text-left outline-none",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-2",
    ],
    backdrop: [
      "fixed inset-0 z-4 bg-sand-12/80",
      "transition-opacity duration-150",
      "data-starting-style:opacity-0 data-ending-style:opacity-0",
    ],
    popup: [
      "fixed inset-0 z-5 m-auto h-fit w-fit outline-none",
      "transition-opacity duration-150",
      "data-starting-style:opacity-0 data-ending-style:opacity-0",
    ],
    title: "sr-only",
    image: [
      "block h-auto w-auto max-h-[calc(100dvh-4rem)] max-w-[calc(100vw-4rem)]",
      "rounded-3 bg-sand-1 shadow-5",
    ],
    close: [
      "fixed top-4 right-4 z-5",
      "inline-flex size-8 cursor-pointer items-center justify-center rounded-full",
      "border border-sand-6 bg-sand-1 text-sand-12 shadow-3",
      "outline-none hover:bg-sand-3",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-2",
      "[&_svg]:size-4",
    ],
  },
});
