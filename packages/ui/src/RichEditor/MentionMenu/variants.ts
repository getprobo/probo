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

export const mentionMenu = tv({
  slots: {
    popup: [
      "z-3 flex max-h-64 w-72 flex-col overflow-y-auto",
      "rounded-3 bg-sand-1 p-2 shadow-5 outline-none",
    ],
    option: [
      "group flex w-full cursor-pointer items-center gap-2 rounded-2 px-2 py-1",
      "text-left outline-none select-none",
      "data-highlighted:bg-gold-9 data-highlighted:text-white",
    ],
    body: "flex min-w-0 flex-1 flex-col",
    label: "block w-full truncate",
    description: "block w-full truncate",
    empty: "px-3 py-1.5",
  },
});
