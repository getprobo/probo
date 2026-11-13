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

import type { TaskFindingComboboxQuery } from "#/__generated__/core/TaskFindingComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useComboboxPage } from "../_lib/useComboboxPage";

const pageSize = 20;

export interface TaskFindingOption {
  id: string;
  title: string;
  referenceId: string;
}

const taskFindingPageQuery = graphql`
  query TaskFindingComboboxQuery(
    $organizationId: ID!
    $first: Int!
    $after: CursorKey
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        findings(
          first: $first
          after: $after
          orderBy: { field: REFERENCE_ID, direction: ASC }
          filter: { query: $query }
        ) {
          edges {
            node {
              id
              title
              referenceId
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

interface TaskFindingComboboxProps {
  excludeIds: ReadonlySet<string>;
  disabled?: boolean;
  onSelect: (finding: TaskFindingOption) => void;
}

export function TaskFindingCombobox({
  excludeIds,
  disabled,
  onSelect,
}: TaskFindingComboboxProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const page = useComboboxPage<TaskFindingComboboxQuery, TaskFindingOption>({
    query: taskFindingPageQuery,
    variables: (after, query) => ({
      organizationId,
      first: pageSize,
      after,
      query: query.trim() || null,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const findings = organization?.findings;
      return {
        items: findings?.edges.map(edge => ({
          id: edge.node.id,
          title: edge.node.title,
          referenceId: edge.node.referenceId,
        })) ?? [],
        hasNextPage: findings?.pageInfo.hasNextPage ?? false,
        endCursor: findings?.pageInfo.endCursor ?? null,
      };
    },
    reloadKey: organizationId,
  });
  const placeholder = t("detailsPage.searchFindings");
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
      ariaLabel={t("detailsPage.findings")}
      placeholder={placeholder}
      emptyLabel={t("detailsPage.noFindingsFound")}
      errorLabel={t("detailsPage.findingsLoadError")}
      loadMoreLabel={t("detailsPage.actions.loadMoreFindings")}
      disabled={disabled}
      onSelect={onSelect}
      itemId={item => item.id}
      itemLabel={item => `${item.referenceId} — ${item.title}`}
    />
  );
}
