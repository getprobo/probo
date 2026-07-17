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

// Data request list: a bordered card of rows, each with a leading type icon, a
// title + reference/message subline, and a trailing time + status badge.
export const rightsRequestList = tv({
  slots: {
    card: "overflow-hidden rounded-4 border border-sand-a4 bg-sand-1",
    item: "flex items-center gap-4 border-b border-sand-a3 px-4 py-3 last:border-b-0",
    icon: "flex size-9 shrink-0 items-center justify-center rounded-3 bg-sand-3 text-sand-a11 [&_svg]:size-4",
    content: "flex min-w-0 flex-1 flex-col gap-0.5",
    subline: "flex min-w-0 items-center gap-1.5",
    trailing: "flex shrink-0 items-center gap-3",
  },
});

// Segmented type selector wrapping under the dialog header, plus the vertical
// stack of form fields.
export const newRequestForm = tv({
  slots: {
    root: "flex flex-col gap-4",
    label: "flex flex-col gap-1.5",
    success: "flex flex-col items-center gap-3 px-2 py-6 text-center",
    successIcon: "flex size-10 items-center justify-center rounded-full bg-green-3 text-green-11 [&_svg]:size-6",
  },
});
