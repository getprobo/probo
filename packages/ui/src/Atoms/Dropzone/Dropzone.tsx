// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { useRefSync } from "@probo/hooks";
import { useCallback } from "react";
import { type FileRejection, useDropzone } from "react-dropzone";
import { useTranslation } from "react-i18next";
import { tv } from "tailwind-variants";

import { IconPageCross, IconUpload } from "../Icons";
import { Spinner } from "../Spinner/Spinner";

type Props = {
  description: string;
  isUploading: boolean;
  disabled?: boolean;
  accept?: Record<string, string[]>;
  maxSize?: number; // maxSize in MB
  onDrop: (acceptedFiles: File[]) => void;
};

export const dropzone = tv({
  slots: {
    wrapper:
            "bg-subtle border border border-border-low p-2 rounded-[20px] outline-none  focus-visible:shadow-focus",
    zone: "bg-secondary min-h-46 border border-border-low border-dashed rounded-2xl flex items-center justify-center",
    title: "flex gap-2 text-sm font-medium text-center justify-center",
    description: "text-xs text-txt-tertiary text-center",
  },
  variants: {
    isDragActive: {
      true: {
        wrapper: "border-border-success shadow-focus",
        zone: "border-border-success",
      },
    },
    hasError: {
      true: {
        wrapper: "bg-danger",
        zone: "bg-invert",
        title: "text-txt-danger",
        description: "text-txt-danger",
      },
    },
    disabled: {
      true: {
        wrapper: "opacity-60 cursor-default",
        zone: "bg-highlight",
      },
      false: {
        zone: "hover:bg-secondary-hover",
      },
    },
  },
  defaultVariants: {
    isDragActive: false,
    disabled: false,
    hasError: false,
  },
});

const MB = 1024 * 1024;

export function Dropzone(props: Props) {
  const { t } = useTranslation();
  const onDropRef = useRefSync(props.onDrop);
  const onDrop = useCallback(
    (files: File[]) => {
      onDropRef.current(files);
    },
    [onDropRef],
  );
  const { getRootProps, getInputProps, isDragActive, fileRejections }
    = useDropzone({
      disabled: props.disabled || props.isUploading,
      accept: props.accept,
      onDrop,
      maxSize: props.maxSize ? props.maxSize * MB : undefined,
    });
  const error = getDropzoneError(t, fileRejections);
  const { wrapper, zone, title, description } = dropzone({
    ...props,
    isDragActive,
    hasError: !!error,
  });

  return (
    <div {...getRootProps()} className={wrapper()}>
      <div className={zone()}>
        <input {...getInputProps({ max: 10 })} />
        {props.isUploading
          ? (
              <div className={title({ isDragActive: true })}>
                <Spinner />
                {t("ui.dropzone.uploading")}
              </div>
            )
          : (
              <div className="space-y-2">
                <div className={title()}>
                  {error
                    ? (
                        <IconPageCross size={20} />
                      )
                    : (
                        <IconUpload size={20} />
                      )}
                  {error ?? t("ui.dropzone.browse")}
                </div>
                <div className={description()}>{props.description}</div>
              </div>
            )}
      </div>
    </div>
  );
}

function getDropzoneError(
  t: (key: string) => string,
  fileRejections: readonly FileRejection[],
) {
  if (fileRejections.length === 0) {
    return null;
  }

  const code = fileRejections[0].errors[0].code;
  switch (code) {
    case "file-invalid-type":
      return t("ui.dropzone.errors.unsupported");
    case "file-too-large":
      return t("ui.dropzone.errors.tooLarge");
    default:
      return t("ui.dropzone.errors.generic");
  }
}
