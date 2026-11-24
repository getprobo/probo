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

import { type Editor, useEditorState } from "@tiptap/react";
import { useEffect, useRef, useState } from "react";

import { linkEditorPluginKey, openLink } from "../LinkExtension";

import { applyLink } from "./formatting";

// useLinkEditor holds the link field's state. It opens when asked, or when a
// link in the text is clicked while editing: the link extension then selects
// it and posts an editor request.
export function useLinkEditor(editor: Editor) {
  const [open, setOpenState] = useState(false);
  const [url, setUrl] = useState("");
  // Closing without submitting moves focus out of the field, and its blur
  // handler submits. This makes that submit a no-op.
  const discarded = useRef(false);
  const state = useEditorState({
    editor,
    selector: ({ editor: current }) => {
      if (current.isDestroyed) {
        return { empty: true, isLink: false, href: undefined, request: null };
      }

      return {
        empty: current.state.selection.empty,
        isLink: current.isActive("link"),
        href: current.getAttributes("link").href as string | undefined,
        request: linkEditorPluginKey.getState(current.state) ?? null,
      };
    },
  });
  const requestId = state.request?.id ?? 0;
  const [handledRequestId, setHandledRequestId] = useState(requestId);

  if (requestId !== handledRequestId) {
    setHandledRequestId(requestId);
    setUrl(state.request?.href ?? "");
    setOpenState(true);
  }

  useEffect(() => {
    if (open) {
      discarded.current = false;
    }
  }, [open]);

  function setOpen(next: boolean) {
    if (next) {
      setUrl(state.href ?? "");
    } else {
      discarded.current = true;
    }
    setOpenState(next);
  }

  // Closes the field and reports whether a new address was applied.
  function submit(): boolean {
    const next = url.trim();
    setOpenState(false);
    if (discarded.current) {
      return false;
    }
    if (next && next !== (state.href ?? "")) {
      applyLink(editor, next);
      return true;
    }
    return false;
  }

  function remove() {
    setOpenState(false);
    applyLink(editor, null);
  }

  function openUrl() {
    openLink(url.trim());
  }

  return {
    open,
    setOpen,
    url,
    setUrl,
    isLink: state.isLink,
    href: state.href,
    empty: state.empty,
    submit,
    remove,
    openUrl,
  };
}
