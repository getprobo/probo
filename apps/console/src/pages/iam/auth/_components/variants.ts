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

// Shared /auth card chrome. The wash matches the employee-portal get-started
// cards: a lime radial pinned to the top of the white card, behind the title.
export const authLayout = tv({
  slots: {
    column: "flex w-full max-w-lg flex-col gap-6",
    header: "relative flex min-h-6 items-center justify-center",
    back: "absolute left-0 top-1/2 -translate-y-1/2",
    frame: "relative w-full overflow-hidden",
    wash: [
      "pointer-events-none absolute inset-x-0 top-0 z-0 h-3/5",
      "bg-[radial-gradient(ellipse_85%_55%_at_50%_0%,rgb(230_255_3/0.72)_0%,rgb(230_255_3/0.28)_35%,transparent_62%)]",
      "mask-[linear-gradient(to_bottom,black_0%,black_40%,transparent_100%)]",
    ],
    body: "relative z-1 px-8 pt-14 pb-8",
  },
});
