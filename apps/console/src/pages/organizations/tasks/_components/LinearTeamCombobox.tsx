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

import { UsersIcon } from "@phosphor-icons/react";
import { ComboboxPaged } from "@probo/ui/src/v2/Combobox/ComboboxPaged";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { LinearTeamComboboxQuery } from "#/__generated__/core/LinearTeamComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { linearTeamLabel } from "../_lib/linearTeamLabel";
import { useComboboxPage } from "../_lib/useComboboxPage";

const PAGE_SIZE = 20;
const NO_TEAM_ID = "__none__";

const teamPageQuery = graphql`
  query LinearTeamComboboxQuery(
    $organizationId: ID!
    $first: Int!
    $after: String
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        linearTeams(first: $first, after: $after, query: $query) {
          edges {
            node {
              id
              name
              key
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

export interface LinearTeamOption {
  id: string;
  name: string;
  key: string;
}

export interface LinearTeamComboboxProps {
  selectedId: string | null;
  selectedLabel: string;
  ariaLabel: string;
  placeholder: string;
  disabled?: boolean;
  noneLabel?: string;
  onSelect: (team: LinearTeamOption) => void;
  onClear?: () => void;
}

export function LinearTeamCombobox({
  selectedId,
  selectedLabel,
  ariaLabel,
  placeholder,
  disabled,
  noneLabel,
  onSelect,
  onClear,
}: LinearTeamComboboxProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const showNone = noneLabel != null && onClear != null;
  const displayId = selectedId ?? (showNone ? NO_TEAM_ID : null);
  const displayLabel = selectedId == null && showNone && noneLabel != null
    ? noneLabel
    : selectedLabel;
  const page = useComboboxPage<LinearTeamComboboxQuery, LinearTeamOption>({
    query: teamPageQuery,
    variables: (after, query) => ({
      organizationId,
      first: PAGE_SIZE,
      after,
      query: query || null,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const teams = organization?.linearTeams;
      return {
        items: teams?.edges.map(edge => edge.node) ?? [],
        hasNextPage: teams?.pageInfo.hasNextPage ?? false,
        endCursor: teams?.pageInfo.endCursor ?? null,
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
      emptyLabel={t("detailsPage.linear.noResults")}
      errorLabel={t("detailsPage.linear.errors.load")}
      loadMoreLabel={t("detailsPage.linear.loadMore")}
      disabled={disabled}
      pinned={noneLabel != null && onClear != null
        ? { id: NO_TEAM_ID, label: noneLabel }
        : undefined}
      onPinnedSelect={onClear}
      onSelect={onSelect}
      itemId={team => team.id}
      itemLabel={team => linearTeamLabel({ name: team.name, key: team.key })}
      itemLeading={() => <UsersIcon />}
    />
  );
}
