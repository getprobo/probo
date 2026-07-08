// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery, useRefetchableFragment } from "react-relay";

import { PageHeader } from "#/components/PageHeader/PageHeader";

import { SubprocessorCategorySection } from "./_components/SubprocessorCategorySection";
import type { SubprocessorNode } from "./_components/SubprocessorCategorySection";
import { SubprocessorsEmpty } from "./_components/SubprocessorsEmpty";
import { SubprocessorsToolbar } from "./_components/SubprocessorsToolbar";
import { groupByCategory } from "./_lib/groupByCategory";
import { toQueryVariables } from "./_lib/toQueryVariables";
import { useSubprocessorFilters } from "./_lib/useSubprocessorFilters";
import type { SubprocessorsPageQuery } from "./__generated__/SubprocessorsPageQuery.graphql";
import type { SubprocessorsPageRefetchQuery } from "./__generated__/SubprocessorsPageRefetchQuery.graphql";
import type { SubprocessorsPage_query$key } from "./__generated__/SubprocessorsPage_query.graphql";

export const subprocessorsPageQuery = graphql`
  query SubprocessorsPageQuery($query: String, $category: SubprocessorCategory, $country: CountryCode) {
    ...SubprocessorsPage_query @arguments(query: $query, category: $category, country: $country)
    ...SubprocessorsToolbar_query
  }
`;

const subprocessorsPageFragment = graphql`
  fragment SubprocessorsPage_query on Query
  @refetchable(queryName: "SubprocessorsPageRefetchQuery")
  @argumentDefinitions(
    query: { type: "String" }
    category: { type: "SubprocessorCategory" }
    country: { type: "CountryCode" }
  ) {
    currentTrustCenter @required(action: THROW) {
      subprocessors(first: 250, filter: { query: $query, category: $category, country: $country }) {
        totalCount
        edges {
          node {
            id
            category
            ...SubprocessorListItem_subprocessor
          }
        }
      }
    }
  }
`;

interface SubprocessorsPageProps {
  queryRef: PreloadedQuery<SubprocessorsPageQuery>;
}

export function SubprocessorsPage({ queryRef }: SubprocessorsPageProps) {
  const { t } = useTranslation("subprocessors");
  const root = usePreloadedQuery<SubprocessorsPageQuery>(subprocessorsPageQuery, queryRef);
  const [data, refetch] = useRefetchableFragment<SubprocessorsPageRefetchQuery, SubprocessorsPage_query$key>(
    subprocessorsPageFragment,
    root,
  );

  const filters = useSubprocessorFilters();
  const { query, category, country } = filters;
  const [isRefetching, startTransition] = useTransition();

  // The initial query already loaded with the URL's filter values; only refetch
  // on subsequent filter changes. Refetch inside a transition so the toolbar and
  // current results stay mounted (no whole-page Suspense fallback) while the
  // filtered results load — the results are just dimmed via `isRefetching`.
  const isFirstRender = useRef(true);
  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }
    startTransition(() => {
      refetch(toQueryVariables({ query, category, country }), { fetchPolicy: "store-or-network" });
    });
  }, [refetch, query, category, country]);

  const { subprocessors } = data.currentTrustCenter;
  const nodes: SubprocessorNode[] = subprocessors.edges.map(edge => edge.node);
  const groups = groupByCategory(nodes, value => t(`categories.${value}.label`));

  return (
    <>
      <PageHeader title={t("title")} count={subprocessors.totalCount}>
        <SubprocessorsToolbar queryKey={root} />
      </PageHeader>
      <div className="flex w-full flex-col items-center px-8 py-8">
        <div
          aria-busy={isRefetching}
          className={`flex w-full max-w-5xl flex-col gap-8 transition-opacity duration-150 ${isRefetching ? "opacity-60" : ""}`}
        >
          {groups.length === 0
            ? <SubprocessorsEmpty />
            : groups.map(group => (
                <SubprocessorCategorySection
                  key={group.category}
                  category={group.category}
                  subprocessors={group.nodes}
                />
              ))}
        </div>
      </div>
    </>
  );
}
