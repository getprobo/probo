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

import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";
import { useParams } from "react-router";

import type { useUploadTaskPictureMutation } from "#/__generated__/core/useUploadTaskPictureMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { taskPicturesConnectionId } from "./taskPath";

const uploadTaskPictureMutation = graphql`
  mutation useUploadTaskPictureMutation(
    $input: UploadTaskPictureInput!
    $connections: [ID!]!
  ) {
    uploadTaskPicture(input: $input) {
      taskPictureEdge @appendEdge(connections: $connections) {
        node {
          id
          ...TaskPictureListItem_taskPicture
        }
      }
    }
  }
`;

export function useUploadTaskPicture() {
  const { t } = useTranslation("organizations/tasks");
  const { taskId } = useParams<{ taskId: string }>();
  const [uploadTaskPicture, isUploading]
    = useMutation<useUploadTaskPictureMutation>(
      uploadTaskPictureMutation,
      {
        successMessage: t("detailsPage.pictures.messages.uploaded"),
        errorToast: t("detailsPage.pictures.errors.upload"),
      },
    );

  async function uploadPicture(file: File) {
    if (taskId == null) {
      throw new Error(":taskId missing in route params");
    }

    await uploadTaskPicture({
      variables: {
        input: {
          taskId,
          file: null,
        },
        connections: [taskPicturesConnectionId(taskId)],
      },
      uploadables: {
        "input.file": file,
      },
    });
  }

  return [uploadPicture, isUploading] as const;
}
