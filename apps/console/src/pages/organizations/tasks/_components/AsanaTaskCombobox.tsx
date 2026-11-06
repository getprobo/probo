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

import { CheckSquareIcon, PlusIcon } from "@phosphor-icons/react";
import { ComboboxPaged } from "@probo/ui/src/v2/Combobox/ComboboxPaged";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { AsanaTaskComboboxQuery } from "#/__generated__/core/AsanaTaskComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useComboboxPage } from "../_lib/useComboboxPage";

const pageSize = 20;

export const newAsanaTaskId = "__new__";

export interface AsanaTaskOption {
  id: string;
  name: string;
}

const taskPageQuery = graphql`
  query AsanaTaskComboboxQuery(
    $organizationId: ID!
    $projectId: String!
    $first: Int!
    $after: String
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        asanaTasks(projectId: $projectId, first: $first, after: $after, query: $query) {
          edges {
            node {
              id
              name
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

export interface AsanaTaskComboboxProps {
  projectId: string | null;
  selectedId: string | null;
  selectedLabel: string;
  disabled?: boolean;
  onSelect: (task: AsanaTaskOption) => void;
  onNewTask: () => void;
}

export function AsanaTaskCombobox({
  projectId,
  selectedId,
  selectedLabel,
  disabled,
  onSelect,
  onNewTask,
}: AsanaTaskComboboxProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const page = useComboboxPage<AsanaTaskComboboxQuery, AsanaTaskOption>({
    query: taskPageQuery,
    variables: (after, query) => ({
      organizationId,
      projectId: projectId ?? "",
      first: pageSize,
      after,
      query: query || null,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const tasks = organization?.asanaTasks;
      return {
        items: tasks?.edges.map(edge => edge.node) ?? [],
        hasNextPage: tasks?.pageInfo.hasNextPage ?? false,
        endCursor: tasks?.pageInfo.endCursor ?? null,
      };
    },
    reloadKey: `${organizationId}:${projectId ?? ""}`,
    enabled: projectId != null,
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
      selectedId={projectId == null ? null : selectedId}
      selectedLabel={selectedLabel}
      ariaLabel={t("detailsPage.asana.task")}
      placeholder={t("detailsPage.asana.searchTask")}
      emptyLabel={t("detailsPage.asana.noResults")}
      errorLabel={t("detailsPage.asana.errors.load")}
      loadMoreLabel={t("detailsPage.asana.loadMore")}
      disabled={disabled || projectId == null}
      pinned={{
        id: newAsanaTaskId,
        label: t("detailsPage.asana.newTask"),
        leading: <PlusIcon />,
      }}
      onPinnedSelect={onNewTask}
      onSelect={onSelect}
      itemId={task => task.id}
      itemLabel={task => task.name}
      itemLeading={() => <CheckSquareIcon />}
    />
  );
}
