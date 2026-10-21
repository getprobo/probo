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

// The label stays in flow. The hint is painted over it, so a button with no
// icon keeps its width when the modifier is held.
export const shortcutLabelSwap = tv({
  slots: {
    root: "relative inline-flex items-center justify-center",
    label: "transition-opacity duration-150 ease-out",
    hint: "pointer-events-none absolute inset-0 flex items-center justify-center transition-opacity duration-150 ease-out",
  },
  variants: {
    pressed: {
      true: {
        label: "opacity-0",
        hint: "opacity-100",
      },
      false: {
        label: "opacity-100",
        hint: "opacity-0",
      },
    },
  },
  defaultVariants: {
    pressed: false,
  },
});
