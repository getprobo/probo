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

import { usesMetaShortcut } from "@probo/hotkeys";

interface ModEnterEvent {
  key: string;
  repeat: boolean;
  altKey: boolean;
  shiftKey: boolean;
  metaKey: boolean;
  ctrlKey: boolean;
  preventDefault: () => void;
}

function matchesModEnter(event: ModEnterEvent, metaShortcut: boolean) {
  if (event.altKey || event.shiftKey || event.key !== "Enter") {
    return false;
  }
  if (metaShortcut) {
    return event.metaKey && !event.ctrlKey;
  }
  return event.ctrlKey && !event.metaKey;
}

// TipTap's hard break owns Mod-Enter. Returning true claims the shortcut,
// including key repeat, so a held key does not insert line breaks.
export function submitOnModEnter(
  event: ModEnterEvent,
  submit: (() => void) | undefined,
  metaShortcut = usesMetaShortcut(),
) {
  if (submit == null || !matchesModEnter(event, metaShortcut)) {
    return false;
  }

  event.preventDefault();
  if (!event.repeat) {
    submit();
  }
  return true;
}
