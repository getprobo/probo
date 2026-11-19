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

export const richEditorToolbar = tv({
  slots: {
    // The host draws the row's border and background around the controls.
    root: "flex min-w-0 flex-wrap items-center justify-start gap-1",
    separator: "mx-1 my-1.5",
    blockTypeTrigger: "min-w-36",
    // Fills the trigger so the caret sits at the end. Regular weight, like the
    // select triggers, instead of the button's medium weight.
    triggerLabel: "flex-1 text-start font-normal",
    // Muted like the select trigger's caret.
    triggerCaret: "text-sand-a10",
    linkPopup: "flex w-80 items-center gap-1 p-1",
    linkInput: "h-8 min-w-0 flex-1 bg-transparent px-2 text-2 text-sand-12 outline-none placeholder:text-sand-9",
  },
});
