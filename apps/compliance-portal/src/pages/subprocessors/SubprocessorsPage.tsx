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

import { useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery, useRefetchableFragment } from "react-relay";

import { ListErrorBoundary } from "#/components/errors/ListErrorBoundary";
import { PageHeader } from "#/components/PageHeader/PageHeader";

import type { SubprocessorsPage_query$key } from "./__generated__/SubprocessorsPage_query.graphql";
import type { SubprocessorsPageQuery } from "./__generated__/SubprocessorsPageQuery.graphql";
import type { SubprocessorsPageRefetchQuery } from "./__generated__/SubprocessorsPageRefetchQuery.graphql";
import type { SubprocessorNode } from "./_components/SubprocessorCategorySection";
import { SubprocessorCategorySection } from "./_components/SubprocessorCategorySection";
import { SubprocessorsEmpty } from "./_components/SubprocessorsEmpty";
import { SubprocessorsToolbar } from "./_components/SubprocessorsToolbar";
import { groupByCategory } from "./_lib/groupByCategory";
import { toQueryVariables } from "./_lib/toQueryVariables";
import { useSubprocessorFilters } from "./_lib/useSubprocessorFilters";

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
    currentCompliancePortal @required(action: THROW) {
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

  const { subprocessors } = data.currentCompliancePortal;
  const nodes: SubprocessorNode[] = subprocessors.edges.map(edge => edge.node);
  const groups = groupByCategory(nodes, value => t(`categories.${value}.label`));

  return (
    <>
      <PageHeader title={t("title")} count={subprocessors.totalCount} flushBottomSpace>
        <SubprocessorsToolbar queryKey={root} />
      </PageHeader>
      <div className="flex w-full flex-col items-center px-8 py-8 max-md:px-4">
        <div
          aria-busy={isRefetching}
          className={`flex w-full max-w-5xl flex-col gap-8 transition-opacity duration-150 ${isRefetching ? "opacity-60" : ""}`}
        >
          <ListErrorBoundary
            onRetry={done => startTransition(() => {
              refetch(
                toQueryVariables({ query, category, country }),
                { fetchPolicy: "network-only", onComplete: done },
              );
            })}
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
          </ListErrorBoundary>
        </div>
      </div>
    </>
  );
}
