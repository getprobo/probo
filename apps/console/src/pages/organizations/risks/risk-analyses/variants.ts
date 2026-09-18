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

export const riskAnalysisDescriptionField = tv({
  slots: {
    editor: "[&_.tiptap>p:first-child]:mt-0",
  },
  variants: {
    density: {
      form: {
        editor: "min-h-40 py-3! [&_.tiptap]:min-h-full",
      },
      display: {
        editor: "flex-none min-h-0 py-0! pl-0! pr-0! shadow-none bg-transparent [&_.tiptap]:min-h-0",
      },
    },
  },
  defaultVariants: {
    density: "form",
  },
});

export const riskAnalysisDetailSummary = tv({
  slots: {
    body: "flex flex-col",
    meta: [
      "grid grid-cols-2 gap-4 md:grid-cols-3",
      "[:not(:first-child)]:mt-4 [:not(:first-child)]:border-t [:not(:first-child)]:border-border-low [:not(:first-child)]:pt-4",
    ],
    label: "mb-1 text-xs font-semibold text-txt-tertiary",
    value: "text-sm text-txt-primary",
  },
});
