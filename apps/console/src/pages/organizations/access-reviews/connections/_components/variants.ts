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

import { tv } from "tailwind-variants";

export const accessReviewSourceSection = tv({
  slots: {
    root: "flex flex-col gap-3",
    header: "flex items-center gap-2.5",
    title: "text-sm font-medium uppercase tracking-wide text-txt-primary",
    count: "text-sm text-txt-tertiary",
    list: "list-none overflow-hidden rounded-[10px] border border-border-low bg-level-1",
    item: "flex flex-wrap items-center gap-4 border-b border-border-low px-4 py-3 last:border-b-0",
    content: "flex min-w-48 flex-1 flex-col gap-0.5 sm:max-w-64",
    trailing: "ml-auto flex min-w-64 flex-1 flex-wrap items-center justify-end gap-3",
    issue: "flex min-w-0 flex-1 flex-wrap items-center gap-2.5",
    issueIcon: "shrink-0 text-txt-danger",
    issueContent: "min-w-48 flex-1 space-y-0.5",
    issueTitle: "text-sm font-medium text-txt-primary",
    issueDescription: "text-xs leading-5 text-txt-secondary",
  },
});
