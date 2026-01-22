import { useRefSync } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { useCallback } from "react";
import { type FileRejection, useDropzone } from "react-dropzone";
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
  const { __ } = useTranslate();
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
  const error = getDropzoneError(__, fileRejections);
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
                {__("Uploading file")}
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
                  {error ?? __("Drag and drop or browse files")}
                </div>
                <div className={description()}>{props.description}</div>
              </div>
            )}
      </div>
    </div>
  );
}

function getDropzoneError(
  __: (s: string) => string,
  fileRejections: readonly FileRejection[],
) {
  if (fileRejections.length === 0) {
    return null;
  }

  const code = fileRejections[0].errors[0].code;
  switch (code) {
    case "file-invalid-type":
      return __("This file is not supported");
    case "file-too-large":
      return __("This file is too large");
    default:
      return __("Something went wrong");
  }
}
