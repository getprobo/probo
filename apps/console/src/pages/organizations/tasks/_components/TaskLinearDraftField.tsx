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

import { HashIcon, PlusIcon } from "@phosphor-icons/react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskLinearDraftField_organization$key } from "#/__generated__/core/TaskLinearDraftField_organization.graphql";
import type { TaskLinearDraftFieldIssueQuery } from "#/__generated__/core/TaskLinearDraftFieldIssueQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { linearIssueOption } from "../_lib/linearIssueOption";
import { linearTeamLabel } from "../_lib/linearTeamLabel";
import type { TaskPriority, TaskState } from "../_lib/taskState";
import { taskLinearPublishField } from "../variants";

import { LinearTeamCombobox } from "./LinearTeamCombobox";
import { RemoteCombobox } from "./RemoteCombobox/RemoteCombobox";

const PAGE_SIZE = 20;
const NEW_ISSUE_ID = "__new__";

interface LinearIssueOption {
  id: string;
  identifier: string;
  title: string;
  content: string | null;
  state: TaskState | null;
  priority: TaskPriority | null;
}

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

const issuePageQuery = graphql`
  query TaskLinearDraftFieldIssueQuery(
    $organizationId: ID!
    $teamId: String!
    $first: Int!
    $after: String
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        linearIssues(teamId: $teamId, first: $first, after: $after, query: $query) {
          edges {
            node {
              id
              identifier
              title
              content
              state
              priority
            }
          }
          pageInfo {
            hasNextPage
            endCursor
          }
        }
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
  const organizationId = useOrganizationId();
  const [override, setOverride] = useState<LinearDraftOverride | null>(null);
  const { root } = taskLinearPublishField();
  const newIssueLabel = t("detailsPage.linear.newIssue");
  const teamId = override === null ? (defaultTeam?.id ?? null) : override.teamId;
  const teamLabel = override === null
    ? (defaultTeam == null
        ? null
        : linearTeamLabel({ name: defaultTeam.name, key: defaultTeam.key }))
    : override.teamLabel;
  const issueId = override?.issueId ?? NEW_ISSUE_ID;
  const issueLabel = override?.issueLabel ?? null;

  function emit(nextTeamId: string | null, nextIssueId: string) {
    onChange({
      teamId: nextTeamId,
      issueId: nextIssueId === NEW_ISSUE_ID ? null : nextIssueId,
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
            issueId: NEW_ISSUE_ID,
            issueLabel: null,
          });
          emit(null, NEW_ISSUE_ID);
        }}
        onSelect={(team) => {
          setOverride({
            teamId: team.id,
            teamLabel: linearTeamLabel({ name: team.name, key: team.key }),
            issueId: NEW_ISSUE_ID,
            issueLabel: null,
          });
          emit(team.id, NEW_ISSUE_ID);
        }}
      />
      <RemoteCombobox<TaskLinearDraftFieldIssueQuery, LinearIssueOption>
        query={issuePageQuery}
        variables={(after, query) => ({
          organizationId,
          teamId: teamId ?? "",
          first: PAGE_SIZE,
          after,
          query: query || null,
        })}
        readPage={(data) => {
          const organization = data.node?.__typename === "Organization" ? data.node : null;
          const page = organization?.linearIssues;
          return {
            items: page?.edges.map(edge => linearIssueOption({
              id: edge.node.id,
              identifier: edge.node.identifier,
              title: edge.node.title,
              content: edge.node.content,
              state: edge.node.state,
              priority: edge.node.priority,
            })) ?? [],
            hasNextPage: page?.pageInfo.hasNextPage ?? false,
            endCursor: page?.pageInfo.endCursor ?? null,
          };
        }}
        reloadKey={`${organizationId}:${teamId ?? ""}`}
        enabled={teamId != null}
        selectedId={teamId == null ? null : issueId}
        selectedLabel={
          teamId == null
            ? t("detailsPage.linear.chooseIssue")
            : issueId === NEW_ISSUE_ID
              ? newIssueLabel
              : (issueLabel ?? newIssueLabel)
        }
        ariaLabel={t("detailsPage.linear.issue")}
        placeholder={t("detailsPage.linear.searchIssue")}
        emptyLabel={t("detailsPage.linear.noResults")}
        disabled={disabled || teamId == null}
        pinned={{
          id: NEW_ISSUE_ID,
          label: newIssueLabel,
        }}
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
        onPinnedSelect={() => {
          setOverride({
            teamId,
            teamLabel,
            issueId: NEW_ISSUE_ID,
            issueLabel: null,
          });
          emit(teamId, NEW_ISSUE_ID);
        }}
        itemId={issue => issue.id}
        itemLabel={issue => `${issue.identifier} ${issue.title}`}
        itemLeading={() => <HashIcon />}
        pinnedLeading={<PlusIcon />}
      />
    </div>
  );
}
