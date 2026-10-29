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

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskAsanaDraftField_organization$key } from "#/__generated__/core/TaskAsanaDraftField_organization.graphql";

import { taskAsanaPublishField } from "../variants";

import { AsanaProjectCombobox } from "./AsanaProjectCombobox";
import { AsanaTaskCombobox, newAsanaTaskId } from "./AsanaTaskCombobox";

export interface AsanaDraft {
  projectId: string | null;
  taskGid: string | null;
}

const taskAsanaDraftFieldFragment = graphql`
  fragment TaskAsanaDraftField_organization on Organization {
    asanaSettings {
      defaultProject {
        id
        name
      }
    }
  }
`;

interface AsanaDraftOverride {
  projectId: string | null;
  projectLabel: string | null;
  taskGid: string;
  taskLabel: string | null;
}

interface TaskAsanaDraftFieldProps {
  organizationKey: TaskAsanaDraftField_organization$key;
  disabled?: boolean;
  onChange: (selection: AsanaDraft) => void;
  onTask?: (task: { name: string }) => void;
}

export function TaskAsanaDraftField({
  organizationKey,
  disabled,
  onChange,
  onTask,
}: TaskAsanaDraftFieldProps) {
  const organization = useFragment(taskAsanaDraftFieldFragment, organizationKey);
  const defaultProject = organization.asanaSettings.defaultProject;
  const { t } = useTranslation("organizations/tasks");
  const [override, setOverride] = useState<AsanaDraftOverride | null>(null);
  const { root } = taskAsanaPublishField();
  const newTaskLabel = t("detailsPage.asana.newTask");
  const projectId = override === null ? (defaultProject?.id ?? null) : override.projectId;
  const projectLabel = override === null
    ? (defaultProject?.name ?? null)
    : override.projectLabel;
  const taskGid = override?.taskGid ?? newAsanaTaskId;
  const taskLabel = override?.taskLabel ?? null;

  function emit(nextProjectId: string | null, nextTaskGid: string) {
    onChange({
      projectId: nextProjectId,
      taskGid: nextTaskGid === newAsanaTaskId ? null : nextTaskGid,
    });
  }

  return (
    <div className={root()}>
      <AsanaProjectCombobox
        selectedId={projectId}
        selectedLabel={projectLabel ?? t("detailsPage.asana.chooseProject")}
        ariaLabel={t("detailsPage.asana.project")}
        placeholder={t("detailsPage.asana.searchProject")}
        disabled={disabled}
        noneLabel={t("detailsPage.asana.noProject")}
        onClear={() => {
          setOverride({
            projectId: null,
            projectLabel: null,
            taskGid: newAsanaTaskId,
            taskLabel: null,
          });
          emit(null, newAsanaTaskId);
        }}
        onSelect={(project) => {
          setOverride({
            projectId: project.id,
            projectLabel: project.name,
            taskGid: newAsanaTaskId,
            taskLabel: null,
          });
          emit(project.id, newAsanaTaskId);
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
        disabled={disabled}
        onSelect={(task) => {
          setOverride({
            projectId,
            projectLabel,
            taskGid: task.id,
            taskLabel: task.name,
          });
          emit(projectId, task.id);
          onTask?.({ name: task.name });
        }}
        onNewTask={() => {
          setOverride({
            projectId,
            projectLabel,
            taskGid: newAsanaTaskId,
            taskLabel: null,
          });
          emit(projectId, newAsanaTaskId);
        }}
      />
    </div>
  );
}
