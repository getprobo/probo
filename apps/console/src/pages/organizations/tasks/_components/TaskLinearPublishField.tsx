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

import type { TaskLinearPublishField_task$key } from "#/__generated__/core/TaskLinearPublishField_task.graphql";
import type { TaskLinearPublishFieldLinkMutation } from "#/__generated__/core/TaskLinearPublishFieldLinkMutation.graphql";
import type { TaskLinearPublishFieldMutation } from "#/__generated__/core/TaskLinearPublishFieldMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { linearTeamLabel } from "../_lib/linearTeamLabel";
import { taskLinearPublishField } from "../variants";

import { LinearIssueCombobox, newLinearIssueId } from "./LinearIssueCombobox";
import { LinearTeamCombobox } from "./LinearTeamCombobox";

const taskLinearPublishFieldFragment = graphql`
  fragment TaskLinearPublishField_task on Task {
    id
  }
`;

const publishMutation = graphql`
  mutation TaskLinearPublishFieldMutation($input: PublishTaskToLinearInput!) {
    publishTaskToLinear(input: $input) {
      task {
        ...TaskLinearField_task
        ...TaskListItem_task
      }
    }
  }
`;

const linkMutation = graphql`
  mutation TaskLinearPublishFieldLinkMutation($input: LinkTaskToLinearInput!) {
    linkTaskToLinear(input: $input) {
      task {
        ...TaskLinearField_task
        ...TaskDetailsPage_task
        ...TaskListItem_task
      }
    }
  }
`;

interface TaskLinearPublishFieldProps {
  taskKey: TaskLinearPublishField_task$key;
  onCompleted?: () => void;
  spacing?: "compact" | "relaxed";
}

export function TaskLinearPublishField({
  taskKey,
  onCompleted,
  spacing = "compact",
}: TaskLinearPublishFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskLinearPublishFieldFragment, taskKey);
  const [teamId, setTeamId] = useState<string | null>(null);
  const [teamLabel, setTeamLabel] = useState<string | null>(null);
  const [issueId, setIssueId] = useState(newLinearIssueId);
  const [issueLabel, setIssueLabel] = useState<string | null>(null);
  const [publish, isPublishing] = useMutation<TaskLinearPublishFieldMutation>(
    publishMutation,
    {
      successMessage: t("detailsPage.linear.published"),
      errorToast: t("detailsPage.linear.errors.publish"),
    },
  );
  const [link, isLinking] = useMutation<TaskLinearPublishFieldLinkMutation>(
    linkMutation,
    {
      successMessage: t("detailsPage.linear.linked"),
      errorToast: t("detailsPage.linear.errors.link"),
    },
  );
  const { root } = taskLinearPublishField({ spacing });
  const isSaving = isPublishing || isLinking;
  const linking = issueId !== newLinearIssueId;
  const newIssueLabel = t("detailsPage.linear.newIssue");

  return (
    <div className={root()}>
      <LinearTeamCombobox
        selectedId={teamId}
        selectedLabel={teamLabel ?? t("detailsPage.linear.chooseTeam")}
        ariaLabel={t("detailsPage.linear.team")}
        placeholder={t("detailsPage.linear.searchTeam")}
        disabled={isSaving}
        onSelect={(team) => {
          setTeamId(team.id);
          setTeamLabel(linearTeamLabel(team));
          setIssueId(newLinearIssueId);
          setIssueLabel(null);
        }}
      />
      <LinearIssueCombobox
        teamId={teamId}
        selectedId={issueId}
        selectedLabel={
          teamId == null
            ? t("detailsPage.linear.chooseIssue")
            : issueId === newLinearIssueId
              ? newIssueLabel
              : (issueLabel ?? newIssueLabel)
        }
        disabled={isSaving}
        onSelect={(issue) => {
          setIssueId(issue.id);
          setIssueLabel(`${issue.identifier} ${issue.title}`);
        }}
        onNewIssue={() => {
          setIssueId(newLinearIssueId);
          setIssueLabel(null);
        }}
      />
      <Button
        className="w-full"
        color="linear"
        disabled={teamId == null || isSaving || (linking && issueId === newLinearIssueId)}
        onClick={() => {
          if (teamId == null) {
            return;
          }

          const request = linking
            ? link({
                variables: {
                  input: {
                    taskId: task.id,
                    teamId,
                    issueId,
                  },
                },
              })
            : publish({
                variables: {
                  input: {
                    taskId: task.id,
                    teamId,
                  },
                },
              });

          void request.then(
            () => {
              onCompleted?.();
            },
            () => {
              // The Linear toast already explains the failure.
            },
          );
        }}
      >
        {linking
          ? t("detailsPage.linear.link")
          : t("detailsPage.linear.publish")}
      </Button>
    </div>
  );
}
