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

import { TrashIcon } from "@phosphor-icons/react";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskPictureListItem_taskPicture$key } from "#/__generated__/core/TaskPictureListItem_taskPicture.graphql";

import { taskPicturesSection } from "../variants";

import { TaskPictureDeleteDialog } from "./TaskPictureDeleteDialog";

const taskPictureListItemFragment = graphql`
  fragment TaskPictureListItem_taskPicture on TaskPicture {
    id
    linearAssetUrl
    canDelete: permission(action: "core:task-picture:delete")
    file {
      fileName
      downloadUrl
    }
    ...TaskPictureDeleteDialog_taskPicture
  }
`;

interface TaskPictureListItemProps {
  taskPictureKey: TaskPictureListItem_taskPicture$key;
}

export function TaskPictureListItem({ taskPictureKey }: TaskPictureListItemProps) {
  const { t } = useTranslation("organizations/tasks");
  const picture = useFragment(taskPictureListItemFragment, taskPictureKey);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { item, image, meta, name } = taskPicturesSection();

  return (
    <li className={item()}>
      <img
        className={image()}
        src={picture.file.downloadUrl}
        alt={picture.file.fileName}
      />
      <div className={meta()}>
        <Text className={name()} size={2} weight="medium">
          {picture.file.fileName}
        </Text>
        {picture.linearAssetUrl != null && (
          <Text size={1} color="faint">
            {t("detailsPage.pictures.attached")}
          </Text>
        )}
      </div>
      {picture.canDelete && (
        <IconButton
          variant="ghost"
          color="neutral"
          aria-label={t("detailsPage.pictures.actions.delete")}
          onClick={() => setDeleteOpen(true)}
        >
          <TrashIcon />
        </IconButton>
      )}
      {picture.canDelete && (
        <TaskPictureDeleteDialog
          taskPictureKey={picture}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
        />
      )}
    </li>
  );
}
