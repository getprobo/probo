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

import { ImageIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { useParams } from "react-router";

import type { TaskPicturesList_pictures$key } from "#/__generated__/core/TaskPicturesList_pictures.graphql";
import type { TaskPicturesListQuery } from "#/__generated__/core/TaskPicturesListQuery.graphql";
import type { TaskPicturesListRefetchQuery } from "#/__generated__/core/TaskPicturesListRefetchQuery.graphql";

import { useUploadTaskPicture } from "../_lib/useUploadTaskPicture";
import { taskPicturesSection } from "../variants";

import { TaskPictureListItem } from "./TaskPictureListItem";

const maxPictureBytes = 10 * 1024 * 1024;

const taskPicturesListQuery = graphql`
  query TaskPicturesListQuery($taskId: ID!) {
    node(id: $taskId) {
      __typename
      ... on Task {
        ...TaskPicturesList_pictures
      }
    }
  }
`;

const taskPicturesListFragment = graphql`
  fragment TaskPicturesList_pictures on Task
  @refetchable(queryName: "TaskPicturesListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    canCreatePicture: permission(action: "core:task-picture:create")
    pictures(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: ASC }
    ) @connection(key: "TaskPicturesSection_pictures") {
      edges {
        node {
          id
          ...TaskPictureListItem_taskPicture
        }
      }
    }
  }
`;

function isPictureFile(file: File) {
  if (file.type === "image/jpeg" || file.type === "image/png" || file.type === "image/webp") {
    return true;
  }

  return /\.(jpe?g|png|webp)$/i.test(file.name);
}

interface TaskPicturesListContentProps {
  taskKey: TaskPicturesList_pictures$key;
}

function TaskPicturesListContent({ taskKey }: TaskPicturesListContentProps) {
  const { t } = useTranslation("organizations/tasks");
  const {
    data: task,
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment<
    TaskPicturesListRefetchQuery,
    TaskPicturesList_pictures$key
  >(taskPicturesListFragment, taskKey);
  const [uploadPicture, isUploading] = useUploadTaskPicture();
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const { root, header, list, actions, fileInput } = taskPicturesSection();
  const pictures = task.pictures.edges.map(edge => edge.node);

  function handleFile(file: File | undefined) {
    if (file == null) {
      return;
    }

    if (!isPictureFile(file)) {
      setError(t("detailsPage.pictures.errors.type"));
      return;
    }

    if (file.size > maxPictureBytes) {
      setError(t("detailsPage.pictures.errors.size"));
      return;
    }

    setError(null);
    void uploadPicture(file).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <section className={root()}>
      <div className={header()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("detailsPage.pictures.title")}
        </Heading>
        {task.canCreatePicture && (
          <Button
            variant="soft"
            color="neutral"
            iconStart={<ImageIcon />}
            loading={isUploading}
            onClick={() => inputRef.current?.click()}
          >
            {t("detailsPage.pictures.actions.upload")}
          </Button>
        )}
      </div>
      {task.canCreatePicture && (
        <input
          ref={inputRef}
          className={fileInput()}
          type="file"
          accept="image/jpeg,image/png,image/webp,.jpg,.jpeg,.png,.webp"
          aria-label={t("detailsPage.pictures.actions.upload")}
          onChange={(event) => {
            const file = event.currentTarget.files?.[0];
            event.currentTarget.value = "";
            handleFile(file);
          }}
        />
      )}
      {error != null && (
        <Text size={2} color="red">
          {error}
        </Text>
      )}
      {pictures.length === 0
        ? (
            <Text size={2} color="faint">
              {t("detailsPage.pictures.empty")}
            </Text>
          )
        : (
            <ul className={list()}>
              {pictures.map(picture => (
                <TaskPictureListItem
                  key={picture.id}
                  taskPictureKey={picture}
                />
              ))}
            </ul>
          )}
      {hasNext && (
        <div className={actions()}>
          <Button
            variant="ghost"
            color="neutral"
            disabled={isLoadingNext}
            onClick={() => {
              loadNext(20);
            }}
          >
            {t("detailsPage.pictures.actions.showMore")}
          </Button>
        </div>
      )}
    </section>
  );
}

function TaskPicturesListLoaded({ taskId }: { taskId: string }) {
  const data = useLazyLoadQuery<TaskPicturesListQuery>(
    taskPicturesListQuery,
    { taskId },
  );

  if (data.node?.__typename !== "Task") {
    return null;
  }

  return <TaskPicturesListContent taskKey={data.node} />;
}

export function TaskPicturesList() {
  const { taskId } = useParams<{ taskId: string }>();
  if (taskId == null) {
    throw new Error(":taskId missing in route params");
  }

  return <TaskPicturesListLoaded taskId={taskId} />;
}
