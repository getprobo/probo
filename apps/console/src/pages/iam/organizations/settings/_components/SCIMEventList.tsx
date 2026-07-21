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

import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { SCIMEventListFragment$key } from "#/__generated__/iam/SCIMEventListFragment.graphql";
import type { SCIMEventListPaginationQuery } from "#/__generated__/iam/SCIMEventListPaginationQuery.graphql";
import { SortableTable } from "#/components/SortableTable";

import { SCIMEventListItem } from "./SCIMEventListItem";

const SCIMEventListFragment = graphql`
  fragment SCIMEventListFragment on SCIMConfiguration
  @refetchable(queryName: "SCIMEventListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    events(first: $first, after: $after, last: $last, before: $before)
      @connection(key: "SCIMEventListFragment_events") {
      edges {
        node {
          id
          ...SCIMEventListItemFragment
        }
      }
    }
  }
`;

export function SCIMEventList(props: { fKey: SCIMEventListFragment$key }) {
  const { fKey } = props;
  const { t } = useTranslation();

  const eventsPagination = usePaginationFragment<
    SCIMEventListPaginationQuery,
    SCIMEventListFragment$key
  >(SCIMEventListFragment, fKey);

  return (
    <SortableTable
      {...eventsPagination}
      refetch={() => {
        eventsPagination.refetch({}, { fetchPolicy: "network-only" });
      }}
      pageSize={20}
    >
      <Thead>
        <Tr>
          <Th>{t("scimEventList.columns.time")}</Th>
          <Th>{t("scimEventList.columns.method")}</Th>
          <Th>{t("scimEventList.columns.path")}</Th>
          <Th>{t("scimEventList.columns.result")}</Th>
        </Tr>
      </Thead>
      <Tbody>
        {!eventsPagination.data.events?.edges
          || eventsPagination.data.events.edges.length === 0
          ? (
              <Tr>
                <Td colSpan={4} className="text-center text-txt-secondary">
                  {t("scimEventList.empty")}
                </Td>
              </Tr>
            )
          : (
              eventsPagination.data.events.edges.map(({ node: event }) => (
                <SCIMEventListItem key={event.id} fKey={event} />
              ))
            )}
      </Tbody>
    </SortableTable>
  );
}
