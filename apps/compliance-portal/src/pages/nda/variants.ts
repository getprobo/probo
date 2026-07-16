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

// Self-contained NDA gate, laid out like the document viewer: a header band
// with the title, subtitle, consent, and sign action above a grey PDF stage.
export const ndaPage = tv({
  slots: {
    root: "flex h-dvh flex-col",
    header: "flex w-full flex-col gap-3",
    text: "flex flex-col gap-1",
    toolbar: "flex min-h-16 items-center justify-between gap-4",
    toolbarStart: "flex items-center gap-2",
    controls: "flex items-center gap-1",
    separator: "h-6",
    consent: "max-w-2xl",
    actions: "flex shrink-0 items-center gap-2",
    body: "min-h-0 flex-1",
    stage: "grid h-full place-items-center bg-sand-3",
  },
});
