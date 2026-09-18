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

import { type Accept, useDropzone } from "react-dropzone";

export type FileDropzoneError = "invalidFileType" | "fileTooLarge";

export interface UseFileDropzoneOptions {
  disabled?: boolean;
  accept: Accept;
  maxSize: number;
  noClick?: boolean;
  noKeyboard?: boolean;
  onFile: (file: File) => void;
  onReject: (error: FileDropzoneError) => void;
}

export function useFileDropzone({
  disabled = false,
  accept,
  maxSize,
  noClick = false,
  noKeyboard = false,
  onFile,
  onReject,
}: UseFileDropzoneOptions) {
  const { getRootProps, getInputProps, isDragActive, open } = useDropzone({
    noClick,
    noKeyboard,
    multiple: false,
    disabled,
    accept,
    maxSize,
    onDrop(acceptedFiles, fileRejections) {
      const rejection = fileRejections[0];
      if (rejection != null) {
        const code = rejection.errors[0]?.code;
        onReject(code === "file-too-large" ? "fileTooLarge" : "invalidFileType");
        return;
      }

      const file = acceptedFiles[0];
      if (file != null) {
        onFile(file);
      }
    },
  });

  return { getRootProps, getInputProps, isDragActive, open };
}
