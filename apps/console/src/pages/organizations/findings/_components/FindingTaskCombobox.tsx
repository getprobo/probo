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

import { ComboboxPaged } from "@probo/ui/src/v2/Combobox/ComboboxPaged";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { FindingTaskComboboxQuery } from "#/__generated__/core/FindingTaskComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useComboboxPage } from "#/pages/organizations/tasks/_lib/useComboboxPage";

const pageSize = 20;

export interface FindingTaskOption {
  id: string;
  name: string;
}

const findingTaskPageQuery = graphql`
  query FindingTaskComboboxQuery(
    $organizationId: ID!
    $first: Int!
    $after: CursorKey
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        tasks(
          first: $first
          after: $after
          orderBy: { field: CREATED_AT, direction: DESC }
          filter: { query: $query }
        ) {
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

interface FindingTaskComboboxProps {
  excludeIds: ReadonlySet<string>;
  disabled?: boolean;
  onSelect: (task: FindingTaskOption) => void;
}

export function FindingTaskCombobox({
  excludeIds,
  disabled,
  onSelect,
}: FindingTaskComboboxProps) {
  const { t } = useTranslation("organizations/findings");
  const organizationId = useOrganizationId();
  const page = useComboboxPage<FindingTaskComboboxQuery, FindingTaskOption>({
    query: findingTaskPageQuery,
    variables: (after, query) => ({
      organizationId,
      first: pageSize,
      after,
      query: query.trim() || null,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const tasks = organization?.tasks;
      return {
        items: tasks?.edges.map(edge => ({
          id: edge.node.id,
          name: edge.node.name,
        })) ?? [],
        hasNextPage: tasks?.pageInfo.hasNextPage ?? false,
        endCursor: tasks?.pageInfo.endCursor ?? null,
      };
    },
    reloadKey: organizationId,
  });
  const placeholder = t("details.tasks.search");
  const items = page.items.filter(item => !excludeIds.has(item.id));

  return (
    <ComboboxPaged
      items={items}
      loading={page.loading}
      failed={page.failed}
      query={page.query}
      onQueryChange={page.onQueryChange}
      hasNext={page.hasNext}
      onLoadMore={page.onLoadMore}
      selectedId={null}
      selectedLabel={placeholder}
      ariaLabel={t("details.tasks.title")}
      placeholder={placeholder}
      emptyLabel={t("details.tasks.emptySearch")}
      errorLabel={t("details.tasks.loadError")}
      loadMoreLabel={t("details.tasks.loadMore")}
      disabled={disabled}
      onSelect={onSelect}
      itemId={item => item.id}
      itemLabel={item => item.name}
    />
  );
}
