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
import { type TaskPriority, type TaskState } from "@probo/tasks";
import { ComboboxPaged } from "@probo/ui/src/v2/Combobox/ComboboxPaged";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { LinearIssueComboboxQuery } from "#/__generated__/core/LinearIssueComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { linearIssueOption } from "../_lib/linearIssueOption";
import { useComboboxPage } from "../_lib/useComboboxPage";

const pageSize = 20;

export const newLinearIssueId = "__new__";

export interface LinearIssueOption {
  id: string;
  identifier: string;
  title: string;
  content: string | null;
  state: TaskState | null;
  priority: TaskPriority | null;
}

const issuePageQuery = graphql`
  query LinearIssueComboboxQuery(
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

export interface LinearIssueComboboxProps {
  teamId: string | null;
  selectedId: string | null;
  selectedLabel: string;
  disabled?: boolean;
  onSelect: (issue: LinearIssueOption) => void;
  onNewIssue: () => void;
}

export function LinearIssueCombobox({
  teamId,
  selectedId,
  selectedLabel,
  disabled,
  onSelect,
  onNewIssue,
}: LinearIssueComboboxProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const page = useComboboxPage<LinearIssueComboboxQuery, LinearIssueOption>({
    query: issuePageQuery,
    variables: (after, query) => ({
      organizationId,
      teamId: teamId ?? "",
      first: pageSize,
      after,
      query: query || null,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const issues = organization?.linearIssues;
      return {
        items: issues?.edges.map(edge => linearIssueOption({
          id: edge.node.id,
          identifier: edge.node.identifier,
          title: edge.node.title,
          content: edge.node.content,
          state: edge.node.state,
          priority: edge.node.priority,
        })) ?? [],
        hasNextPage: issues?.pageInfo.hasNextPage ?? false,
        endCursor: issues?.pageInfo.endCursor ?? null,
      };
    },
    reloadKey: `${organizationId}:${teamId ?? ""}`,
    enabled: teamId != null,
  });

  return (
    <ComboboxPaged
      items={page.items}
      loading={page.loading}
      failed={page.failed}
      query={page.query}
      onQueryChange={page.onQueryChange}
      hasNext={page.hasNext}
      onLoadMore={page.onLoadMore}
      selectedId={teamId == null ? null : selectedId}
      selectedLabel={selectedLabel}
      ariaLabel={t("detailsPage.linear.issue")}
      placeholder={t("detailsPage.linear.searchIssue")}
      emptyLabel={t("detailsPage.linear.noResults")}
      errorLabel={t("detailsPage.linear.errors.load")}
      loadMoreLabel={t("detailsPage.linear.loadMore")}
      disabled={disabled || teamId == null}
      pinned={{
        id: newLinearIssueId,
        label: t("detailsPage.linear.newIssue"),
        leading: <PlusIcon />,
      }}
      onPinnedSelect={onNewIssue}
      onSelect={onSelect}
      itemId={issue => issue.id}
      itemLabel={issue => `${issue.identifier} ${issue.title}`}
      itemLeading={() => <HashIcon />}
    />
  );
}
