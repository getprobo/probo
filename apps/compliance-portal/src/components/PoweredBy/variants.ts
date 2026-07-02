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

// "Powered by Probo" footer (Figma "Powered by"): a borderless, full-width row
// on the body surface, with a dotted texture that fades in from the bottom.
export const poweredBy = tv({
  slots: {
    root: "relative flex w-full items-center justify-center overflow-hidden bg-sand-2 py-6 text-sand-11",
    backdrop: "pointer-events-none absolute inset-0",
    backdropFade: "pointer-events-none absolute inset-0 bg-linear-to-b from-sand-2 to-sand-2/0",
    content: "relative z-10 flex items-center justify-center gap-2",
    logo: "h-6 w-auto text-sand-9",
  },
});
