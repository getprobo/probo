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

import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import { PageHeader } from "#/components/PageHeader/PageHeader";

import { SubprocessorCategorySection } from "./_components/SubprocessorCategorySection";
import type { SubprocessorNode } from "./_components/SubprocessorCategorySection";
import { SubprocessorsEmpty } from "./_components/SubprocessorsEmpty";
import { groupByCategory } from "./_lib/groupByCategory";
import type { SubprocessorsPageQuery } from "./__generated__/SubprocessorsPageQuery.graphql";

export const subprocessorsPageQuery = graphql`
  query SubprocessorsPageQuery {
    currentTrustCenter @required(action: THROW) {
      subprocessors(first: 250) {
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
  const data = usePreloadedQuery<SubprocessorsPageQuery>(subprocessorsPageQuery, queryRef);
  const { subprocessors } = data.currentTrustCenter;

  const nodes: SubprocessorNode[] = subprocessors.edges.map(edge => edge.node);
  const groups = groupByCategory(nodes, category => t(`categories.${category}.label`));

  return (
    <>
      <PageHeader title={t("title")} count={subprocessors.totalCount} />
      <div className="flex w-full flex-col items-center px-8 py-8">
        <div className="flex w-full max-w-5xl flex-col gap-8">
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
