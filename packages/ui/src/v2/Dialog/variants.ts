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

// Modal dialog (Radix "Dialog" over Base UI). The popup is centered and gets
// vertical padding + a 16px gap; each region (header, body, footer) carries its
// own horizontal padding so a full-bleed body slot stays possible.
export const dialog = tv({
  slots: {
    backdrop: [
      "fixed inset-0 z-50 bg-sand-12/40",
      "transition-opacity duration-150",
      "data-starting-style:opacity-0 data-ending-style:opacity-0",
    ],
    popup: [
      "fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2",
      "flex w-[calc(100vw-2rem)] max-w-[600px] flex-col gap-4",
      "max-h-[calc(100vh-2rem)] overflow-y-auto overflow-x-clip",
      "rounded-5 border border-sand-6 bg-sand-1 py-6 shadow-6 outline-none",
      "transition-all duration-150",
      "data-starting-style:scale-95 data-starting-style:opacity-0",
      "data-ending-style:scale-95 data-ending-style:opacity-0",
    ],
    header: "flex flex-col gap-2 px-6",
    title: "text-4 font-medium text-sand-12",
    description: "text-2 text-sand-11",
    body: "px-6",
    footer: "flex flex-wrap items-center justify-end gap-3 px-6 max-sm:flex-col-reverse max-sm:items-stretch",
  },
});

// Static frame matching the dialog popup, without the interactive positioning,
// so a placeholder can render before Base UI (and the content) loads.
export const dialogSkeleton = tv({
  base: [
    "flex w-full max-w-[600px] flex-col gap-4",
    "rounded-5 border border-sand-6 bg-sand-1 py-6 shadow-6",
  ],
});
