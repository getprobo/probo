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

import { FolderIcon } from "@phosphor-icons/react";
import { ComboboxPaged } from "@probo/ui/src/v2/Combobox/ComboboxPaged";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { AsanaProjectComboboxQuery } from "#/__generated__/core/AsanaProjectComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useComboboxPage } from "../_lib/useComboboxPage";

const PAGE_SIZE = 20;
const NO_PROJECT_ID = "__none__";

const projectPageQuery = graphql`
  query AsanaProjectComboboxQuery(
    $organizationId: ID!
    $first: Int!
    $after: String
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        asanaProjects(first: $first, after: $after, query: $query) {
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

export interface AsanaProjectOption {
  id: string;
  name: string;
}

export interface AsanaProjectComboboxProps {
  selectedId: string | null;
  selectedLabel: string;
  ariaLabel: string;
  placeholder: string;
  disabled?: boolean;
  noneLabel?: string;
  onSelect: (project: AsanaProjectOption) => void;
  onClear?: () => void;
}

export function AsanaProjectCombobox({
  selectedId,
  selectedLabel,
  ariaLabel,
  placeholder,
  disabled,
  noneLabel,
  onSelect,
  onClear,
}: AsanaProjectComboboxProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const showNone = noneLabel != null && onClear != null;
  const displayId = selectedId ?? (showNone ? NO_PROJECT_ID : null);
  const displayLabel = selectedId == null && showNone && noneLabel != null
    ? noneLabel
    : selectedLabel;
  const page = useComboboxPage<AsanaProjectComboboxQuery, AsanaProjectOption>({
    query: projectPageQuery,
    variables: (after, query) => ({
      organizationId,
      first: PAGE_SIZE,
      after,
      query: query || null,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const projects = organization?.asanaProjects;
      return {
        items: projects?.edges.map(edge => edge.node) ?? [],
        hasNextPage: projects?.pageInfo.hasNextPage ?? false,
        endCursor: projects?.pageInfo.endCursor ?? null,
      };
    },
    reloadKey: organizationId,
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
      selectedId={displayId}
      selectedLabel={displayLabel}
      ariaLabel={ariaLabel}
      placeholder={placeholder}
      emptyLabel={t("detailsPage.asana.noResults")}
      errorLabel={t("detailsPage.asana.errors.load")}
      loadMoreLabel={t("detailsPage.asana.loadMore")}
      disabled={disabled}
      pinned={noneLabel != null && onClear != null
        ? { id: NO_PROJECT_ID, label: noneLabel }
        : undefined}
      onPinnedSelect={onClear}
      onSelect={onSelect}
      itemId={project => project.id}
      itemLabel={project => project.name}
      itemLeading={() => <FolderIcon />}
    />
  );
}
