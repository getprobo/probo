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
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
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
  const [confirmLink, setConfirmLink] = useState(false);
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
        size={1}
        disabled={teamId == null || isSaving || (linking && issueId === newLinearIssueId)}
        onClick={() => {
          if (teamId == null) {
            return;
          }

          if (linking) {
            setConfirmLink(true);
            return;
          }

          void publish({
            variables: {
              input: {
                taskId: task.id,
                teamId,
              },
            },
          }).then(
            () => {
              onCompleted?.();
            },
            () => {
              // The Linear toast already explains the failure.
            },
          );
        }}
      >
        {linking ? t("detailsPage.linear.link") : t("detailsPage.linear.publish")}
      </Button>
      <Dialog open={confirmLink} onOpenChange={setConfirmLink}>
        <DialogPopup>
          <DialogHeader>
            <DialogTitle>{t("detailsPage.linear.confirmLink.title")}</DialogTitle>
            <DialogDescription>
              {t("detailsPage.linear.confirmLink.description")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral" disabled={isSaving}>
                  {t("detailsPage.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="button"
              variant="solid"
              color="neutral"
              highContrast
              loading={isLinking}
              disabled={teamId == null}
              onClick={() => {
                if (teamId == null) {
                  return;
                }

                void link({
                  variables: {
                    input: {
                      taskId: task.id,
                      teamId,
                      issueId,
                    },
                  },
                }).then(
                  () => {
                    setConfirmLink(false);
                    onCompleted?.();
                  },
                  () => {
                    // The Linear toast already explains the failure.
                  },
                );
              }}
            >
              {t("detailsPage.linear.link")}
            </Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>
    </div>
  );
}
