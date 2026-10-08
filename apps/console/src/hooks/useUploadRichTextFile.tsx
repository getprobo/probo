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

import { RichTextUploadContext } from "@probo/ui";
import { type ReactNode, useCallback } from "react";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { useUploadRichTextFileMutation } from "#/__generated__/core/useUploadRichTextFileMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const uploadFileMutation = graphql`
  mutation useUploadRichTextFileMutation($input: UploadAttachmentFileInput!) {
    uploadAttachmentFile(input: $input) {
      file {
        id
        fileName
        mimeType
        size
      }
    }
  }
`;

export function useUploadRichTextFile() {
  const { organizationId } = useParams();
  const [mutate] = useMutation<useUploadRichTextFileMutation>(
    uploadFileMutation,
    { errorToast: false },
  );

  return useCallback(async (file: File) => {
    if (!organizationId) {
      throw new Error("missing organization");
    }

    const response = await mutate({
      variables: {
        input: {
          organizationId,
          file: null,
        },
      },
      uploadables: {
        "input.file": file,
      },
    });

    const uploaded = response.uploadAttachmentFile?.file;
    if (!uploaded) {
      throw new Error("upload failed");
    }

    return {
      id: uploaded.id,
      fileName: uploaded.fileName,
      mimeType: uploaded.mimeType,
      size: uploaded.size,
    };
  }, [mutate, organizationId]);
}

export function RichTextUploads({ children }: { children: ReactNode }) {
  const upload = useUploadRichTextFile();

  return (
    <RichTextUploadContext value={upload}>
      {children}
    </RichTextUploadContext>
  );
}
