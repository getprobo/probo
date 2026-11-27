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

import { dateTimeFormat } from "@probo/i18n";
import { Button } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { AssetChangeHistory_asset$key } from "#/__generated__/core/AssetChangeHistory_asset.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

const PAGE_SIZE = 20;

const assetChangeHistoryFragment = graphql`
  fragment AssetChangeHistory_asset on Asset
  @refetchable(queryName: "AssetChangeHistoryRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    changeEvents(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: OBSERVED_AT, direction: DESC }
    ) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          changeType
          before
          after
          observedAt
          assetSyncId
          actor {
            fullName
          }
        }
      }
    }
  }
`;

function formatChangeJSON(value: string | null | undefined) {
  if (value == null || value === "") {
    return "";
  }

  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
}

export function AssetChangeHistory({
  asset,
}: {
  asset: AssetChangeHistory_asset$key;
}) {
  const { t, i18n } = useTranslation();
  const [data, refetch] = useRefetchableFragment(
    assetChangeHistoryFragment,
    asset,
  );
  const connection = data.changeEvents;
  const { isPending, goPrevious, goNext } = useCursorPagination(
    (variables: CursorPaginationVariables) => {
      refetch(variables, { fetchPolicy: "store-or-network" });
    },
    connection.pageInfo,
    PAGE_SIZE,
  );
  const events = connection.edges.map(edge => edge.node);

  return (
    <section className="space-y-3 max-w-2xl" aria-busy={isPending}>
      <h2 className="text-lg font-medium">
        {t("assetDetailsPage.changeHistory.title")}
      </h2>
      {events.length === 0
        ? (
            <p className="text-sm text-txt-tertiary">
              {t("assetDetailsPage.changeHistory.empty")}
            </p>
          )
        : (
            <ol className="space-y-4">
              {events.map(event => {
                const actor = event.actor?.fullName
                  ?? (event.assetSyncId
                    ? t("assetDetailsPage.changeHistory.sync")
                    : "");
                const before = formatChangeJSON(event.before);
                const after = formatChangeJSON(event.after);

                return (
                  <li key={event.id} className="space-y-2 border-t border-border-subtle pt-4">
                    <div className="flex items-baseline justify-between gap-4">
                      <span className="font-medium">
                        {t(`assetDetailsPage.changeHistory.types.${event.changeType}`)}
                      </span>
                      <span className="text-sm text-txt-tertiary">
                        {dateTimeFormat(i18n.language, event.observedAt)}
                      </span>
                    </div>
                    {actor !== "" && (
                      <p className="text-sm text-txt-secondary">{actor}</p>
                    )}
                    {before !== "" && (
                      <pre className="overflow-x-auto text-xs">
                        {t("assetDetailsPage.changeHistory.before")}
                        {"\n"}
                        {before}
                      </pre>
                    )}
                    {after !== "" && (
                      <pre className="overflow-x-auto text-xs">
                        {t("assetDetailsPage.changeHistory.after")}
                        {"\n"}
                        {after}
                      </pre>
                    )}
                  </li>
                );
              })}
            </ol>
          )}
      {(connection.pageInfo.hasPreviousPage || connection.pageInfo.hasNextPage) && (
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            disabled={!connection.pageInfo.hasPreviousPage || isPending}
            onClick={goPrevious}
          >
            {t("assetDetailsPage.changeHistory.previous")}
          </Button>
          <Button
            type="button"
            variant="secondary"
            disabled={!connection.pageInfo.hasNextPage || isPending}
            onClick={goNext}
          >
            {t("assetDetailsPage.changeHistory.next")}
          </Button>
        </div>
      )}
    </section>
  );
}
