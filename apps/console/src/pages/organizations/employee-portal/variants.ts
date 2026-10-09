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

export const employeePortalPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex flex-col gap-2",
  },
});

export const employeePortalPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex flex-col gap-2",
    section: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
  },
});

export const employeePortalBrandingSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    body: "flex flex-col gap-4",
    logos: "flex items-start gap-4 max-sm:flex-col",
    logoCell: "flex min-w-0 flex-1 flex-col gap-2",
    darkIsland: "dark scheme-dark rounded-4 bg-sand-1 p-2",
  },
});

export const employeePortalCapabilitiesSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
  },
});
