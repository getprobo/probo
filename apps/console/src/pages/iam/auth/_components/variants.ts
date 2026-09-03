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

// Shared /auth card chrome. A lime circle sits behind the card (wider than
// the card) so the glow peeks past the left and right edges.
export const authLayout = tv({
  slots: {
    column: "flex w-full max-w-lg flex-col gap-6",
    header: "relative z-2 flex min-h-6 items-center justify-center",
    back: "absolute left-0 top-1/2 -translate-y-1/2",
    stack: "relative",
    wash: [
      "pointer-events-none absolute left-1/2 top-1/2 z-0 aspect-square w-[calc(140%+32px)] -translate-x-1/2 -translate-y-1/2",
      "bg-[radial-gradient(circle,rgb(230_255_3/0.95)_0%,rgb(230_255_3/0.5)_40%,transparent_68%)]",
    ],
    frame: "relative z-1 w-full",
    body: "relative z-1 px-8 pt-14 pb-8",
  },
});
