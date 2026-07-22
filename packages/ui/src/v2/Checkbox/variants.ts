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

// A 20px control box (Radix "Checkbox"). The checked/indeterminate surface and
// the disabled treatment resolve off Base UI's data-* state attributes.
export const checkbox = tv({
  base: [
    "inline-flex size-5 shrink-0 cursor-pointer items-center justify-center rounded-2 border border-sand-a7 bg-sand-1 text-gold-1 outline-none transition-colors",
    "focus-visible:ring-2 focus-visible:ring-gold-8",
    "data-[checked]:border-gold-12 data-[checked]:bg-gold-12",
    "data-[indeterminate]:border-gold-12 data-[indeterminate]:bg-gold-12",
    "data-[disabled]:cursor-not-allowed data-[disabled]:border-sand-a3 data-[disabled]:bg-sand-2",
  ],
});

export const checkboxIndicator = tv({
  base: "flex items-center justify-center text-current [&_svg]:size-3.5",
});

export const checkboxSkeleton = tv({
  base: "inline-block size-5 shrink-0 animate-pulse rounded-2 bg-sand-3 align-middle",
});
