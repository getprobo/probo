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

import { ImageIcon, TrashIcon } from "@phosphor-icons/react";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { MouseEvent } from "react";

import type { FileDropzoneError } from "#/lib/useFileDropzone";
import { useFileDropzone } from "#/lib/useFileDropzone";

import { imageDropzone } from "./variants";

export const imageDropzoneAccept = {
  "image/jpeg": [".jpg", ".jpeg"],
  "image/png": [".png"],
  "image/svg+xml": [".svg"],
  "image/webp": [".webp"],
};

export const imageDropzoneMaxBytes = 5 * 1024 * 1024;

export interface ImageDropzoneProps {
  src?: string | null;
  ratio?: "square" | "wide";
  disabled?: boolean;
  uploading?: boolean;
  placeholder: string;
  clearLabel?: string;
  onFile: (file: File) => void;
  onClear?: () => void;
  onReject: (error: FileDropzoneError) => void;
}

function DropzoneCopy({ className, label }: { className: string; label: string }) {
  return (
    <div className={className}>
      <ImageIcon size={24} weight="duotone" className="text-sand-9" />
      <Text size={1} color="neutral">
        {label}
      </Text>
    </div>
  );
}

export function ImageDropzone({
  src,
  ratio = "square",
  disabled = false,
  uploading = false,
  placeholder,
  clearLabel,
  onFile,
  onClear,
  onReject,
}: ImageDropzoneProps) {
  const filled = src != null && src.length > 0;
  const { getRootProps, getInputProps, isDragActive } = useFileDropzone({
    disabled: disabled || uploading,
    accept: imageDropzoneAccept,
    maxSize: imageDropzoneMaxBytes,
    onFile,
    onReject,
  });
  const { root, image, placeholder: placeholderSlot, overlay, hint, clear } = imageDropzone({
    ratio,
    filled,
    dragActive: isDragActive,
    disabled: disabled || uploading,
  });

  function handleClear(event: MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    event.stopPropagation();
    onClear?.();
  }

  const showHint = filled && !disabled && !uploading && !isDragActive;

  return (
    <div {...getRootProps({ "className": root(), "aria-label": placeholder })}>
      <input {...getInputProps()} />
      {filled
        ? (
            <img src={src ?? undefined} alt="" className={image()} />
          )
        : (
            <DropzoneCopy className={placeholderSlot()} label={placeholder} />
          )}
      {showHint && (
        <DropzoneCopy className={hint()} label={placeholder} />
      )}
      {uploading && (
        <div className={overlay()}>
          <Spinner size={2} />
        </div>
      )}
      {filled && onClear != null && !disabled && !uploading && (
        <div className={clear()}>
          <IconButton
            type="button"
            size={1}
            variant="outline"
            color="red"
            aria-label={clearLabel ?? placeholder}
            disabled={uploading}
            onClick={handleClear}
          >
            <TrashIcon />
          </IconButton>
        </div>
      )}
    </div>
  );
}
