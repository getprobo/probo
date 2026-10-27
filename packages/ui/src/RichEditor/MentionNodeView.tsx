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

import { PreviewCard } from "@base-ui/react/preview-card";
import { NodeViewWrapper, type ReactNodeViewProps } from "@tiptap/react";
import { useCallback, useContext, useEffect, useRef, useState } from "react";

import { overlayPortalContainer, OverlayPortalRootContext } from "../lib/overlayPortalRoot";

import { getMentionStorage, type MentionItem } from "./MentionExtension";
import { MentionPerson } from "./MentionMenu/MentionPerson";
import { mentionMenu } from "./MentionMenu/variants";

const mentionPreviewDelay = 300;

export function MentionNodeView({ editor, node }: ReactNodeViewProps) {
  const id = typeof node.attrs.id === "string" ? node.attrs.id : "";
  const storedLabel = typeof node.attrs.label === "string" ? node.attrs.label : "";
  const label = storedLabel !== "" ? storedLabel : id;
  const [resolved, setResolved] = useState<MentionItem | null>(null);
  const overlayRoot = useContext(OverlayPortalRootContext);
  const { popup, row } = mentionMenu();
  const item = resolved?.id === id ? resolved : { id, label };
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => () => abortRef.current?.abort(), []);

  const load = useCallback(() => {
    const resolve = getMentionStorage(editor)?.source?.resolve;
    if (resolve == null || id === "") {
      return;
    }

    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    void resolve(id, controller.signal).then((next) => {
      if (controller.signal.aborted || next == null) {
        return;
      }

      setResolved(next);
    }, () => {});
  }, [editor, id]);

  return (
    <NodeViewWrapper
      as="span"
      style={{ display: "inline", whiteSpace: "nowrap" }}
    >
      <PreviewCard.Root>
        <PreviewCard.Trigger
          render={<span />}
          className="mention"
          delay={mentionPreviewDelay}
          contentEditable={false}
          data-type="mention"
          data-id={id}
          data-label={label}
          onPointerEnter={load}
        >
          {`@${label}`}
        </PreviewCard.Trigger>
        <PreviewCard.Portal container={overlayPortalContainer(undefined, overlayRoot)}>
          <PreviewCard.Positioner
            className="z-3"
            side="bottom"
            align="start"
            sideOffset={4}
          >
            <PreviewCard.Popup
              className={popup()}
              data-rich-editor-floating=""
            >
              <div className={row()}>
                <MentionPerson item={item} />
              </div>
            </PreviewCard.Popup>
          </PreviewCard.Positioner>
        </PreviewCard.Portal>
      </PreviewCard.Root>
    </NodeViewWrapper>
  );
}
