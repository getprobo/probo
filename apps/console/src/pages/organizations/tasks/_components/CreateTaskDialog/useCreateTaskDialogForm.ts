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

import { clampUtf8 } from "@probo/helpers";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { useCreateTaskDialogFormLinkMutation } from "#/__generated__/core/useCreateTaskDialogFormLinkMutation.graphql";
import type { useCreateTaskDialogFormPublishMutation } from "#/__generated__/core/useCreateTaskDialogFormPublishMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { isRichEditorContentEmpty } from "#/pages/organizations/_lib/richEditorContent";

import { taskNameMaxLength } from "../../_lib/taskNameMaxLength";
import type { TaskPriority, TaskState } from "../../_lib/taskState";
import { useCreateTask } from "../../_lib/useCreateTask";
import type { LinearDraft, LinearIssueDraft } from "../TaskLinearDraftField";

const publishMutation = graphql`
  mutation useCreateTaskDialogFormPublishMutation($input: PublishTaskToLinearInput!) {
    publishTaskToLinear(input: $input) {
      task {
        id
      }
    }
  }
`;

const linkMutation = graphql`
  mutation useCreateTaskDialogFormLinkMutation($input: LinkTaskToLinearInput!) {
    linkTaskToLinear(input: $input) {
      task {
        ...TaskDetailsPage_task
        ...TaskListItem_task
      }
    }
  }
`;

export interface CreateTaskDialogSessionProps {
  connectionId?: string;
  internalControlId?: string;
  onCompleted?: () => void;
  onClose: () => void;
}

export interface CreateTaskDialogFormState {
  name: string;
  content: string;
  editorKey: number;
  errors: Record<string, string>;
  state: TaskState;
  priority: TaskPriority;
  linear: LinearDraft | null;
  uploading: boolean;
  isSaving: boolean;
  setName: (value: string) => void;
  setContent: (value: string) => void;
  setErrors: (update: (current: Record<string, string>) => Record<string, string>) => void;
  setState: (state: TaskState) => void;
  setPriority: (priority: TaskPriority) => void;
  setLinear: (selection: LinearDraft) => void;
  setUploading: (uploading: boolean) => void;
  applyLinearIssue: (issue: LinearIssueDraft) => void;
  reset: () => void;
  submit: (selection: LinearDraft) => void;
}

export function useCreateTaskDialogForm({
  connectionId,
  internalControlId,
  onCompleted,
  onClose,
}: CreateTaskDialogSessionProps): CreateTaskDialogFormState {
  const { t } = useTranslation("organizations/tasks");
  const [name, setName] = useState("");
  const [content, setContent] = useState("");
  const [uploading, setUploading] = useState(false);
  const [editorKey, setEditorKey] = useState(0);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [state, setState] = useState<TaskState>("TODO");
  const [priority, setPriority] = useState<TaskPriority>("MEDIUM");
  const [createTask, isCreating] = useCreateTask();
  const [linear, setLinearState] = useState<LinearDraft | null>(null);
  const [publish, isPublishing] = useMutation<useCreateTaskDialogFormPublishMutation>(
    publishMutation,
    {
      successMessage: t("detailsPage.linear.published"),
      errorToast: t("detailsPage.linear.errors.publish"),
    },
  );
  const [link, isLinking] = useMutation<useCreateTaskDialogFormLinkMutation>(
    linkMutation,
    {
      successMessage: t("detailsPage.linear.linked"),
      errorToast: t("detailsPage.linear.errors.link"),
    },
  );
  const isSaving = isCreating || isPublishing || isLinking;

  function applyLinearIssue(issue: LinearIssueDraft) {
    setName(clampUtf8(issue.title.trim(), taskNameMaxLength));
    setContent(issue.content ?? "");
    setEditorKey(key => key + 1);
    if (issue.state != null) {
      setState(issue.state);
    }
    if (issue.priority != null) {
      setPriority(issue.priority);
    }
    setErrors({});
  }

  function reset() {
    setName("");
    setContent("");
    setEditorKey(key => key + 1);
    setErrors({});
    setState("TODO");
    setPriority("MEDIUM");
    setLinearState(null);
    setUploading(false);
  }

  function submit(linearSelection: LinearDraft) {
    if (uploading) {
      return;
    }

    const nextName = name.trim();
    if (!nextName) {
      setErrors({ name: t("createDialog.errors.nameRequired") });
      return;
    }
    setErrors({});

    void createTask(
      {
        name: nextName,
        content: isRichEditorContentEmpty(content) ? null : content,
        state,
        priority,
        internalControlIds: internalControlId ? [internalControlId] : undefined,
        linearTeamId: null,
      },
      connectionId,
    ).then(
      async (taskId) => {
        try {
          if (linearSelection.teamId != null) {
            if (linearSelection.issueId != null) {
              await link({
                variables: {
                  input: {
                    taskId,
                    teamId: linearSelection.teamId,
                    issueId: linearSelection.issueId,
                  },
                },
              });
            } else {
              await publish({
                variables: {
                  input: {
                    taskId,
                    teamId: linearSelection.teamId,
                  },
                },
              });
            }
          }
        } catch {
          // The task exists; the Linear toast already explains the failure.
        }
        reset();
        onClose();
        onCompleted?.();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return {
    name,
    content,
    editorKey,
    errors,
    state,
    priority,
    linear,
    uploading,
    isSaving,
    setName,
    setContent,
    setErrors,
    setState,
    setPriority,
    setLinear: setLinearState,
    setUploading,
    applyLinearIssue,
    reset,
    submit,
  };
}
