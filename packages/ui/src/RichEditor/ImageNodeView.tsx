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

import { ImageBrokenIcon } from "@phosphor-icons/react";
import type { ReactNodeViewProps } from "@tiptap/react";
import { NodeViewWrapper } from "@tiptap/react";
import {
  type KeyboardEvent as ReactKeyboardEvent,
  type PointerEvent as ReactPointerEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";

import { attachmentPath, useAttachmentParentID } from "./attachments";
import { dragImageWidth, keyboardImageWidth, loadedImagePin, snapImageWidth } from "./imageWidth";

type ResizeSide = "left" | "right";

export function ImageNodeView({ editor, node, getPos, updateAttributes }: ReactNodeViewProps) {
  const { t } = useTranslation();
  const parentId = useAttachmentParentID();
  const fileId = typeof node.attrs.fileId === "string" ? node.attrs.fileId : "";
  const storedSrc = typeof node.attrs.src === "string" ? node.attrs.src.trim() : "";
  const src = fileId ? attachmentPath(fileId, parentId) : storedSrc;
  const alt = typeof node.attrs.alt === "string" ? node.attrs.alt : "";
  const title = typeof node.attrs.title === "string" ? node.attrs.title : undefined;
  const storedWidth = typeof node.attrs.width === "number" ? node.attrs.width : null;
  const [dragWidth, setDragWidth] = useState<number | null>(null);
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const [pinned, setPinned] = useState<{ src: string; width: number } | null>(null);
  const intrinsicWidth = pinned?.src === src ? pinned.width : null;
  const slotRef = useRef<HTMLDivElement>(null);
  const imageRef = useRef<HTMLImageElement>(null);
  const stopDrag = useRef<(() => void) | null>(null);
  const percent = dragWidth ?? storedWidth;

  useEffect(() => () => stopDrag.current?.(), []);

  const applyLoadedImage = useCallback((image: HTMLImageElement) => {
    const pin = loadedImagePin({
      naturalWidth: image.naturalWidth,
      naturalHeight: image.naturalHeight,
      renderedWidth: image.getBoundingClientRect().width,
    });
    if (pin.failed) {
      setFailedSrc(src);
      return;
    }

    if (pin.width != null) {
      setPinned({ src, width: pin.width });
    }
  }, [src]);

  const setImageRef = useCallback((image: HTMLImageElement | null) => {
    imageRef.current = image;
    if (image != null && image.complete) {
      applyLoadedImage(image);
    }
  }, [applyLoadedImage]);

  if (src === "" || failedSrc === src) {
    return (
      <NodeViewWrapper>
        <div className="rich-image-missing" contentEditable={false}>
          <ImageBrokenIcon size={16} weight="bold" />
          <span>{alt !== "" ? alt : t("ui.richEditor.imageUnavailable")}</span>
        </div>
      </NodeViewWrapper>
    );
  }

  const startResize = (side: ResizeSide, event: ReactPointerEvent<HTMLButtonElement>) => {
    const slot = slotRef.current;
    const image = imageRef.current;
    if (slot == null || image == null) {
      return;
    }

    const slotWidth = slot.getBoundingClientRect().width;
    if (slotWidth <= 0) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    event.currentTarget.setPointerCapture(event.pointerId);

    const startX = event.clientX;
    const startPercent = storedWidth
      ?? (image.getBoundingClientRect().width / slotWidth) * 100;
    const targets = otherImageWidths(editor, getPos());
    const live = { current: snapImageWidth(startPercent, targets) };
    const moved = { current: false };

    const apply = (ev: PointerEvent) => {
      if (ev.clientX === startX) {
        return;
      }

      moved.current = true;
      const bounds = slot.getBoundingClientRect();
      live.current = dragImageWidth({
        side,
        clientX: ev.clientX,
        startX,
        startPercent,
        slotLeft: bounds.left,
        slotRight: bounds.right,
        slotWidth: bounds.width,
        targets,
      });
      setDragWidth(live.current);
    };
    const stop = (ev?: Event) => {
      window.removeEventListener("pointermove", apply);
      window.removeEventListener("pointerup", stop);
      window.removeEventListener("pointercancel", stop);
      stopDrag.current = null;

      if (ev instanceof PointerEvent && ev.type !== "pointercancel") {
        apply(ev);
      }

      if (moved.current && live.current !== storedWidth) {
        updateAttributes({ width: live.current });
      }

      setDragWidth(null);
    };

    stopDrag.current?.();
    stopDrag.current = stop;
    window.addEventListener("pointermove", apply);
    window.addEventListener("pointerup", stop);
    window.addEventListener("pointercancel", stop);
  };

  const resizeFromKeyboard = (side: ResizeSide, event: ReactKeyboardEvent<HTMLButtonElement>) => {
    const slot = slotRef.current;
    const image = imageRef.current;
    if (slot == null || image == null) {
      return;
    }

    const slotWidth = slot.getBoundingClientRect().width;
    if (slotWidth <= 0) {
      return;
    }

    const startPercent = storedWidth
      ?? (image.getBoundingClientRect().width / slotWidth) * 100;
    const next = keyboardImageWidth({
      side,
      key: event.key,
      width: startPercent,
      targets: otherImageWidths(editor, getPos()),
    });
    if (next == null) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();

    if (next !== storedWidth) {
      updateAttributes({ width: next });
    }
  };

  return (
    <NodeViewWrapper className="rich-image-slot">
      <div ref={slotRef} className="w-full">
        <div
          className="rich-image-frame"
          style={percent == null ? undefined : { width: `${percent}%` }}
          contentEditable={false}
        >
          <img
            ref={setImageRef}
            className={percent == null ? "rich-image" : "rich-image rich-image-sized"}
            style={percent == null && intrinsicWidth != null ? { width: intrinsicWidth } : undefined}
            src={src}
            alt={alt}
            title={title}
            draggable={false}
            onLoad={event => applyLoadedImage(event.currentTarget)}
            onError={() => setFailedSrc(src)}
          />
          {editor.isEditable && (
            <>
              <button
                type="button"
                className="rich-image-handle rich-image-handle-left"
                aria-label={t("ui.richEditor.resizeImage")}
                draggable={false}
                onDragStart={event => event.preventDefault()}
                onMouseDown={event => event.stopPropagation()}
                onPointerDown={event => startResize("left", event)}
                onKeyDown={event => resizeFromKeyboard("left", event)}
              />
              <button
                type="button"
                className="rich-image-handle rich-image-handle-right"
                aria-label={t("ui.richEditor.resizeImage")}
                draggable={false}
                onDragStart={event => event.preventDefault()}
                onMouseDown={event => event.stopPropagation()}
                onPointerDown={event => startResize("right", event)}
                onKeyDown={event => resizeFromKeyboard("right", event)}
              />
            </>
          )}
        </div>
      </div>
    </NodeViewWrapper>
  );
}

function otherImageWidths(editor: ReactNodeViewProps["editor"], pos: number | undefined): number[] {
  const widths: number[] = [];

  editor.state.doc.descendants((image, imagePos) => {
    if (image.type.name !== "image" || imagePos === pos || typeof image.attrs.width !== "number") {
      return;
    }

    widths.push(image.attrs.width);
  });

  return widths;
}
