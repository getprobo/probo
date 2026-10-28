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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskAsanaPublishField_task$key } from "#/__generated__/core/TaskAsanaPublishField_task.graphql";
import type { TaskAsanaPublishFieldLinkMutation } from "#/__generated__/core/TaskAsanaPublishFieldLinkMutation.graphql";
import type { TaskAsanaPublishFieldMutation } from "#/__generated__/core/TaskAsanaPublishFieldMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { taskAsanaPublishField } from "../variants";

import { AsanaProjectCombobox } from "./AsanaProjectCombobox";
import { AsanaTaskCombobox, newAsanaTaskId } from "./AsanaTaskCombobox";

const taskAsanaPublishFieldFragment = graphql`
  fragment TaskAsanaPublishField_task on Task {
    id
  }
`;

const publishMutation = graphql`
  mutation TaskAsanaPublishFieldMutation($input: PublishTaskToAsanaInput!) {
    publishTaskToAsana(input: $input) {
      task {
        ...TaskAsanaField_task
        ...TaskListItem_task
      }
    }
  }
`;

const linkMutation = graphql`
  mutation TaskAsanaPublishFieldLinkMutation($input: LinkTaskToAsanaInput!) {
    linkTaskToAsana(input: $input) {
      task {
        ...TaskAsanaField_task
        ...TaskDetailsPage_task
        ...TaskListItem_task
      }
    }
  }
`;

interface TaskAsanaPublishFieldProps {
  taskKey: TaskAsanaPublishField_task$key;
  onCompleted?: () => void;
  spacing?: "compact" | "relaxed";
}

export function TaskAsanaPublishField({
  taskKey,
  onCompleted,
  spacing = "compact",
}: TaskAsanaPublishFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskAsanaPublishFieldFragment, taskKey);
  const [projectId, setProjectId] = useState<string | null>(null);
  const [projectLabel, setProjectLabel] = useState<string | null>(null);
  const [taskGid, setTaskGid] = useState(newAsanaTaskId);
  const [taskLabel, setTaskLabel] = useState<string | null>(null);
  const [publish, isPublishing] = useMutation<TaskAsanaPublishFieldMutation>(
    publishMutation,
    {
      successMessage: t("detailsPage.asana.published"),
      errorToast: t("detailsPage.asana.errors.publish"),
    },
  );
  const [link, isLinking] = useMutation<TaskAsanaPublishFieldLinkMutation>(
    linkMutation,
    {
      successMessage: t("detailsPage.asana.linked"),
      errorToast: t("detailsPage.asana.errors.link"),
    },
  );
  const { root, action } = taskAsanaPublishField({ spacing });
  const isSaving = isPublishing || isLinking;
  const linking = taskGid !== newAsanaTaskId;
  const newTaskLabel = t("detailsPage.asana.newTask");

  return (
    <div className={root()}>
      <AsanaProjectCombobox
        selectedId={projectId}
        selectedLabel={projectLabel ?? t("detailsPage.asana.chooseProject")}
        ariaLabel={t("detailsPage.asana.project")}
        placeholder={t("detailsPage.asana.searchProject")}
        disabled={isSaving}
        onSelect={(project) => {
          setProjectId(project.id);
          setProjectLabel(project.name);
          setTaskGid(newAsanaTaskId);
          setTaskLabel(null);
        }}
      />
      <AsanaTaskCombobox
        projectId={projectId}
        selectedId={taskGid}
        selectedLabel={
          projectId == null
            ? t("detailsPage.asana.chooseTask")
            : taskGid === newAsanaTaskId
              ? newTaskLabel
              : (taskLabel ?? newTaskLabel)
        }
        disabled={isSaving}
        onSelect={(remote) => {
          setTaskGid(remote.id);
          setTaskLabel(remote.name);
        }}
        onNewTask={() => {
          setTaskGid(newAsanaTaskId);
          setTaskLabel(null);
        }}
      />
      <Button
        className={action()}
        color="neutral"
        highContrast
        loading={isSaving}
        disabled={projectId == null || isSaving}
        onClick={() => {
          if (projectId == null) {
            return;
          }

          const request = linking
            ? link({
                variables: {
                  input: {
                    taskId: task.id,
                    projectId,
                    taskGid,
                  },
                },
              })
            : publish({
                variables: {
                  input: {
                    taskId: task.id,
                    projectId,
                  },
                },
              });

          void request.then(
            () => {
              onCompleted?.();
            },
            () => {
              // The Asana toast already explains the failure.
            },
          );
        }}
      >
        {linking
          ? t("detailsPage.asana.link")
          : t("detailsPage.asana.publish")}
      </Button>
    </div>
  );
}
