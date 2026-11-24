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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessInventoryGrantChangeEventList_grant$key } from "#/__generated__/core/AccessInventoryGrantChangeEventList_grant.graphql";
import type { AccessInventoryGrantChangeEventListRefetchQuery } from "#/__generated__/core/AccessInventoryGrantChangeEventListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { ACCESS_INVENTORY_GRANT_CHANGE_EVENTS_PAGE_SIZE } from "../_lib/pageSize";

import { AccessInventoryGrantChangeEventListItem } from "./AccessInventoryGrantChangeEventListItem";
import { accessInventoryGrantChangeEventList } from "./variants";

export const accessInventoryGrantChangeEventListFragment = graphql`
  fragment AccessInventoryGrantChangeEventList_grant on AccessInventoryGrant
  @refetchable(queryName: "AccessInventoryGrantChangeEventListRefetchQuery")
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
      orderBy: { direction: DESC, field: EFFECTIVE_AT }
    ) @required(action: THROW) {
      totalCount
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...AccessInventoryGrantChangeEventListItem_event
        }
      }
    }
  }
`;

interface AccessInventoryGrantChangeEventListProps {
  grantKey: AccessInventoryGrantChangeEventList_grant$key;
}

export function AccessInventoryGrantChangeEventList({
  grantKey,
}: AccessInventoryGrantChangeEventListProps) {
  const { t } = useTranslation("organizations/access-reviews");
  const [grant, refetch] = useRefetchableFragment<
    AccessInventoryGrantChangeEventListRefetchQuery,
    AccessInventoryGrantChangeEventList_grant$key
  >(accessInventoryGrantChangeEventListFragment, grantKey);

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    grant.changeEvents.pageInfo,
    ACCESS_INVENTORY_GRANT_CHANGE_EVENTS_PAGE_SIZE,
  );

  const edges = grant.changeEvents.edges;
  const pageInfo = grant.changeEvents.pageInfo;
  const { root, heading, results, empty, item, pager }
    = accessInventoryGrantChangeEventList({ pending: isPending });

  return (
    <div className={root()}>
      <div className={heading()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("grantDetail.changelog.title", {
            count: grant.changeEvents.totalCount,
          })}
        </Heading>
        <Text size={2} color="neutral">
          {t("grantDetail.changelog.description")}
        </Text>
      </div>
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {t("grantDetail.changelog.empty")}
                </Text>
              </div>
            </Card>
          )
        : (
            <>
              <div aria-busy={isPending} className={results()}>
                <List>
                  {edges.map(({ node }) => (
                    <ListItem key={node.id} className={item()}>
                      <AccessInventoryGrantChangeEventListItem
                        eventKey={node}
                      />
                    </ListItem>
                  ))}
                </List>
              </div>
              <div className={pager()}>
                <Pagination
                  hasPrevious={pageInfo.hasPreviousPage}
                  hasNext={pageInfo.hasNextPage}
                  previousLabel={t("grants.actions.previous")}
                  nextLabel={t("grants.actions.next")}
                  showLabels
                  disabled={isPending}
                  onPrevious={goPrevious}
                  onNext={goNext}
                />
              </div>
            </>
          )}
    </div>
  );
}
