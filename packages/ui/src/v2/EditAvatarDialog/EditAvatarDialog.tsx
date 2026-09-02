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

import { avatarFallback } from "@probo/helpers";
import { type ChangeEvent, useRef } from "react";

import { Avatar } from "../Avatar/Avatar";
import { Button } from "../Button/Button";
import { Dialog } from "../Dialog/Dialog";
import { DialogBody } from "../Dialog/DialogBody";
import { DialogClose } from "../Dialog/DialogClose";
import { DialogDescription } from "../Dialog/DialogDescription";
import { DialogFooter } from "../Dialog/DialogFooter";
import { DialogHeader } from "../Dialog/DialogHeader";
import { DialogPopup } from "../Dialog/DialogPopup";
import { DialogTitle } from "../Dialog/DialogTitle";

import { editAvatarDialog } from "./variants";

export const identityAvatarAccept = "image/png,image/jpeg,image/webp";

export interface EditAvatarDialogProps {
  open: boolean;
  fullName: string;
  src?: string | null;
  uploading?: boolean;
  removing?: boolean;
  title: string;
  description: string;
  uploadLabel: string;
  replaceLabel: string;
  removeLabel: string;
  closeLabel: string;
  onOpenChange: (open: boolean) => void;
  onUpload: (file: File) => void;
  onRemove?: () => void;
}

export function EditAvatarDialog({
  open,
  fullName,
  src,
  uploading = false,
  removing = false,
  title,
  description,
  uploadLabel,
  replaceLabel,
  removeLabel,
  closeLabel,
  onOpenChange,
  onUpload,
  onRemove,
}: EditAvatarDialogProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const slots = editAvatarDialog();
  const busy = uploading || removing;
  const fallback = avatarFallback(fullName);
  const previewSrc = src ?? undefined;

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) {
      return;
    }
    onUpload(file);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup className={slots.popup()}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogBody>
          <div className={slots.preview()}>
            <Avatar
              size={7}
              variant="soft"
              color="gold"
              radius="full"
              src={previewSrc}
              alt={fullName}
              fallback={fallback}
            />
            <div className={slots.actions()}>
              <input
                ref={inputRef}
                className={slots.fileInput()}
                type="file"
                accept={identityAvatarAccept}
                disabled={busy}
                onChange={handleChange}
              />
              <Button
                variant="solid"
                color="neutral"
                highContrast
                size={2}
                loading={uploading}
                disabled={busy}
                onClick={() => inputRef.current?.click()}
              >
                {previewSrc ? replaceLabel : uploadLabel}
              </Button>
              {previewSrc && onRemove && (
                <Button
                  variant="soft"
                  color="red"
                  size={2}
                  loading={removing}
                  disabled={busy}
                  onClick={onRemove}
                >
                  {removeLabel}
                </Button>
              )}
            </div>
          </div>
        </DialogBody>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral" highContrast size={2}>
                {closeLabel}
              </Button>
            )}
          />
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
