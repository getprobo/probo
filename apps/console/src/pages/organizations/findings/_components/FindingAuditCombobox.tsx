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

import type { FindingAuditComboboxQuery } from "#/__generated__/core/FindingAuditComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useComboboxPage } from "#/pages/organizations/tasks/_lib/useComboboxPage";

const pageSize = 20;

export interface FindingAuditOption {
  id: string;
  name: string;
}

const findingAuditPageQuery = graphql`
  query FindingAuditComboboxQuery(
    $organizationId: ID!
    $first: Int!
    $after: CursorKey
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        audits(
          first: $first
          after: $after
          orderBy: { field: CREATED_AT, direction: DESC }
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

interface FindingAuditComboboxProps {
  excludeIds: ReadonlySet<string>;
  disabled?: boolean;
  onSelect: (audit: FindingAuditOption) => void;
}

export function FindingAuditCombobox({
  excludeIds,
  disabled,
  onSelect,
}: FindingAuditComboboxProps) {
  const { t } = useTranslation("organizations/findings");
  const organizationId = useOrganizationId();
  const untitled = t("details.audits.untitled");
  const page = useComboboxPage<FindingAuditComboboxQuery, FindingAuditOption>({
    query: findingAuditPageQuery,
    variables: after => ({
      organizationId,
      first: pageSize,
      after,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const audits = organization?.audits;
      return {
        items: audits?.edges.map(edge => ({
          id: edge.node.id,
          name: edge.node.name?.trim() || untitled,
        })) ?? [],
        hasNextPage: audits?.pageInfo.hasNextPage ?? false,
        endCursor: audits?.pageInfo.endCursor ?? null,
      };
    },
    reloadKey: organizationId,
  });
  const needle = page.query.trim().toLowerCase();
  const items = page.items.filter((item) => {
    if (excludeIds.has(item.id)) {
      return false;
    }
    return needle === "" || item.name.toLowerCase().includes(needle);
  });

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
      selectedLabel={t("details.audits.search")}
      ariaLabel={t("details.audits.title")}
      placeholder={t("details.audits.search")}
      emptyLabel={t("details.audits.emptySearch")}
      errorLabel={t("details.audits.loadError")}
      loadMoreLabel={t("details.audits.loadMore")}
      disabled={disabled}
      onSelect={onSelect}
      itemId={item => item.id}
      itemLabel={item => item.name}
    />
  );
}
