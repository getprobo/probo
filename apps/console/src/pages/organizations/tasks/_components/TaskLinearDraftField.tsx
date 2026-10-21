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

import { type TaskPriority, type TaskState } from "@probo/tasks";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskLinearDraftField_organization$key } from "#/__generated__/core/TaskLinearDraftField_organization.graphql";

import { linearTeamLabel } from "../_lib/linearTeamLabel";
import { taskLinearPublishField } from "../variants";

import { LinearIssueCombobox, newLinearIssueId } from "./LinearIssueCombobox";
import { LinearTeamCombobox } from "./LinearTeamCombobox";

export interface LinearIssueDraft {
  title: string;
  content: string | null;
  state: TaskState | null;
  priority: TaskPriority | null;
}

export interface LinearDraft {
  teamId: string | null;
  issueId: string | null;
}

const taskLinearDraftFieldFragment = graphql`
  fragment TaskLinearDraftField_organization on Organization {
    linearSettings {
      defaultTeam {
        id
        name
        key
      }
    }
  }
`;

interface LinearDraftOverride {
  teamId: string | null;
  teamLabel: string | null;
  issueId: string;
  issueLabel: string | null;
}

interface TaskLinearDraftFieldProps {
  organizationKey: TaskLinearDraftField_organization$key;
  disabled?: boolean;
  onChange: (selection: LinearDraft) => void;
  onIssue?: (issue: LinearIssueDraft) => void;
}

export function TaskLinearDraftField({
  organizationKey,
  disabled,
  onChange,
  onIssue,
}: TaskLinearDraftFieldProps) {
  const organization = useFragment(taskLinearDraftFieldFragment, organizationKey);
  const defaultTeam = organization.linearSettings.defaultTeam;
  const { t } = useTranslation("organizations/tasks");
  const [override, setOverride] = useState<LinearDraftOverride | null>(null);
  const { root } = taskLinearPublishField();
  const newIssueLabel = t("detailsPage.linear.newIssue");
  const teamId = override === null ? (defaultTeam?.id ?? null) : override.teamId;
  const teamLabel = override === null
    ? (defaultTeam == null
        ? null
        : linearTeamLabel({ name: defaultTeam.name, key: defaultTeam.key }))
    : override.teamLabel;
  const issueId = override?.issueId ?? newLinearIssueId;
  const issueLabel = override?.issueLabel ?? null;

  function emit(nextTeamId: string | null, nextIssueId: string) {
    onChange({
      teamId: nextTeamId,
      issueId: nextIssueId === newLinearIssueId ? null : nextIssueId,
    });
  }

  return (
    <div className={root()}>
      <LinearTeamCombobox
        selectedId={teamId}
        selectedLabel={teamLabel ?? t("detailsPage.linear.chooseTeam")}
        ariaLabel={t("detailsPage.linear.team")}
        placeholder={t("detailsPage.linear.searchTeam")}
        disabled={disabled}
        noneLabel={t("detailsPage.linear.noTeam")}
        onClear={() => {
          setOverride({
            teamId: null,
            teamLabel: null,
            issueId: newLinearIssueId,
            issueLabel: null,
          });
          emit(null, newLinearIssueId);
        }}
        onSelect={(team) => {
          setOverride({
            teamId: team.id,
            teamLabel: linearTeamLabel({ name: team.name, key: team.key }),
            issueId: newLinearIssueId,
            issueLabel: null,
          });
          emit(team.id, newLinearIssueId);
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
        disabled={disabled}
        onSelect={(issue) => {
          setOverride({
            teamId,
            teamLabel,
            issueId: issue.id,
            issueLabel: `${issue.identifier} ${issue.title}`,
          });
          emit(teamId, issue.id);
          onIssue?.({
            title: issue.title,
            content: issue.content,
            state: issue.state,
            priority: issue.priority,
          });
        }}
        onNewIssue={() => {
          setOverride({
            teamId,
            teamLabel,
            issueId: newLinearIssueId,
            issueLabel: null,
          });
          emit(teamId, newLinearIssueId);
        }}
      />
    </div>
  );
}
