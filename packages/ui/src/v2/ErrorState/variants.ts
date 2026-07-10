// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { tv } from "tailwind-variants/lite";

// Centered full-page error message block for portal boundaries (Figma
// "Error message / Page"). Used for 404 / 403 / 500 / generic page-level errors.
//   root     the centering wrapper (page vs in-shell sizing)
//   block    the 256px content column (code + title + description + actions)
//   content  the stacked text region
//   actions  the primary/secondary action row
export const errorState = tv({
  slots: {
    root: "flex w-full items-center justify-center",
    block: "flex min-w-64 max-w-md flex-col items-center gap-6 text-center",
    content: "flex w-full flex-col items-center gap-2",
    actions: "flex items-center justify-center gap-2",
  },
  variants: {
    // Standalone fills the viewport; in-shell sits inside the app chrome.
    fullPage: {
      true: { root: "min-h-screen px-6 py-16" },
      false: { root: "px-6 py-12" },
    },
  },
  defaultVariants: {
    fullPage: false,
  },
});
