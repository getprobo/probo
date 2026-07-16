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

// Document list: category sections, each a titled block above a bordered list.
// Entries share a common layout (title + accent metadata + trailing access
// action).

export const documentSection = tv({
  slots: {
    root: "flex flex-col gap-3",
    header: "flex flex-col gap-0.5",
    list: "overflow-hidden rounded-4 border border-sand-a4 bg-sand-1",
  },
});

export const documentListItem = tv({
  slots: {
    root: "flex items-center gap-4 border-b border-sand-a3 px-4 py-3 last:border-b-0",
    content: "flex min-w-0 flex-1 flex-col gap-0.5",
  },
});

// PDF preview: a scrollable grey stage holding the stacked, centered pages.
export const pdfPreview = tv({
  slots: {
    viewport: "h-full overflow-auto bg-sand-3",
    list: "flex flex-col items-center gap-4 py-8",
    loading: "grid place-items-center py-16 text-sand-a10",
    spinner: "size-6 animate-spin",
    page: "shadow-3",
  },
});

// Full-page viewer: fixed header band with the title and toolbar above a
// scrollable grey stage for the PDF / image / download-fallback body.
export const documentViewer = tv({
  slots: {
    root: "flex h-full flex-col",
    header: "flex w-full flex-col gap-3",
    back: "-ml-2 self-start",
    toolbar: "flex min-h-16 items-center justify-between gap-4",
    toolbarStart: "flex items-center gap-2",
    controls: "flex items-center gap-1",
    actions: "flex items-center gap-2",
    separator: "h-6",
    body: "min-h-0 flex-1",
    stage: "grid h-full place-items-center bg-sand-3",
    imageStage: "grid h-full place-items-center overflow-auto bg-sand-3 p-8",
    image: "max-h-full max-w-full object-contain shadow-3",
    spinner: "size-6 animate-spin text-sand-a10",
  },
});
