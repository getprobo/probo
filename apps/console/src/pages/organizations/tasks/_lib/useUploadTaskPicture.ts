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

import { isPictureFile, type RichEditorPicture, useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";
import { useParams } from "react-router";

import type { useUploadTaskPictureMutation } from "#/__generated__/core/useUploadTaskPictureMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const maxPictureBytes = 10 * 1024 * 1024;

const uploadTaskPictureMutation = graphql`
  mutation useUploadTaskPictureMutation($input: UploadTaskPictureInput!) {
    uploadTaskPicture(input: $input) {
      taskPictureEdge {
        node {
          id
          file {
            downloadUrl
            fileName
          }
        }
      }
    }
  }
`;

export function useUploadTaskPicture() {
  const { t } = useTranslation("organizations/tasks");
  const { toast } = useToast();
  const { taskId } = useParams<{ taskId: string }>();
  const [uploadTaskPicture] = useMutation<useUploadTaskPictureMutation>(
    uploadTaskPictureMutation,
    {
      errorToast: t("detailsPage.pictures.errors.upload"),
    },
  );

  async function uploadPicture(file: File): Promise<RichEditorPicture> {
    if (taskId == null) {
      throw new Error(":taskId missing in route params");
    }

    if (!isPictureFile(file)) {
      const title = t("detailsPage.pictures.errors.type");
      toast({ title, description: "", variant: "error" });
      throw new Error(title);
    }

    if (file.size > maxPictureBytes) {
      const title = t("detailsPage.pictures.errors.size");
      toast({ title, description: "", variant: "error" });
      throw new Error(title);
    }

    const result = await uploadTaskPicture({
      variables: {
        input: {
          taskId,
          file: null,
        },
      },
      uploadables: {
        "input.file": file,
      },
    });
    const uploaded = result.uploadTaskPicture.taskPictureEdge.node.file;

    return {
      src: uploaded.downloadUrl,
      alt: uploaded.fileName,
    };
  }

  return uploadPicture;
}
