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

import type { DocumentControlComboboxQuery } from "#/__generated__/core/DocumentControlComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useComboboxPage } from "#/pages/organizations/tasks/_lib/useComboboxPage";

const pageSize = 20;

export interface DocumentControlOption {
  id: string;
  sectionTitle: string;
  name: string;
}

const documentControlPageQuery = graphql`
  query DocumentControlComboboxQuery(
    $organizationId: ID!
    $first: Int!
    $after: CursorKey
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        controls(
          first: $first
          after: $after
          orderBy: { field: SECTION_TITLE, direction: ASC }
          filter: { query: $query }
        ) {
          edges {
            node {
              id
              sectionTitle
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

export interface DocumentControlComboboxProps {
  excludeIds: ReadonlySet<string>;
  disabled?: boolean;
  onSelect: (control: DocumentControlOption) => void;
}

function controlLabel(control: Pick<DocumentControlOption, "sectionTitle" | "name">) {
  return `${control.sectionTitle} ${control.name}`;
}

export function DocumentControlCombobox({
  excludeIds,
  disabled,
  onSelect,
}: DocumentControlComboboxProps) {
  const { t } = useTranslation("organizations/documents");
  const organizationId = useOrganizationId();
  const page = useComboboxPage<DocumentControlComboboxQuery, DocumentControlOption>({
    query: documentControlPageQuery,
    variables: (after, query) => ({
      organizationId,
      first: pageSize,
      after,
      query: query.trim() || null,
    }),
    readPage: (data) => {
      const organization = data.node?.__typename === "Organization" ? data.node : null;
      const controls = organization?.controls;
      return {
        items: controls?.edges.map(edge => ({
          id: edge.node.id,
          sectionTitle: edge.node.sectionTitle,
          name: edge.node.name,
        })) ?? [],
        hasNextPage: controls?.pageInfo.hasNextPage ?? false,
        endCursor: controls?.pageInfo.endCursor ?? null,
      };
    },
    reloadKey: organizationId,
  });
  const placeholder = t("details.controls.add");
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
      ariaLabel={t("details.controls.title")}
      placeholder={placeholder}
      emptyLabel={t("details.controls.empty")}
      errorLabel={t("details.controls.loadError")}
      loadMoreLabel={t("details.controls.loadMore")}
      disabled={disabled}
      onSelect={onSelect}
      itemId={item => item.id}
      itemLabel={controlLabel}
    />
  );
}
