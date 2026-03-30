// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Button, IconChevronDown, Spinner, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAccessListFragment$key } from "#/__generated__/core/CompliancePageAccessListFragment.graphql";
import type { CompliancePageAccessListQuery } from "#/__generated__/core/CompliancePageAccessListQuery.graphql";

import { CompliancePageAccessListItem } from "./CompliancePageAccessListItem";

const fragment = graphql`
  fragment CompliancePageAccessListFragment on TrustCenter
  @argumentDefinitions(
    first: { type: Int, defaultValue: 10 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: TrustCenterAccessOrder, defaultValue: { field: CREATED_AT, direction: DESC } }
  )
  @refetchable(queryName: "CompliancePageAccessListQuery") {
    accesses(
      first: $first
      after: $after
      orderBy: $order
    ) @connection(key: "CompliancePageAccessList_accesses" filters: ["orderBy"]) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...CompliancePageAccessListItemFragment
        }
      }
    }
  }
`;

export function CompliancePageAccessList(props: {
  fragmentRef: CompliancePageAccessListFragment$key;
}) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();

  const {
    data: { accesses },
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment<CompliancePageAccessListQuery, CompliancePageAccessListFragment$key>(
    fragment,
    fragmentRef,
  );

  return accesses.edges.length === 0
    ? (
        <Table>
          <Tbody>
            <Tr>
              <Td className="text-center text-txt-tertiary py-8">
                {__("No external access granted yet")}
              </Td>
            </Tr>
          </Tbody>
        </Table>
      )
    : (
        <>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Name")}</Th>
                <Th>{__("Email")}</Th>
                <Th>{__("Date")}</Th>
                <Th className="text-center">{__("Access")}</Th>
                <Th className="text-center">{__("Requests")}</Th>
                <Th className="text-center">{__("NDA")}</Th>
                <Th></Th>
              </Tr>
            </Thead>
            <Tbody>
              {accesses.edges.map(({ node: access }) => (
                <CompliancePageAccessListItem
                  key={access.id}
                  fragmentRef={access}
                />
              ))}
            </Tbody>
          </Table>
          {hasNext && (
            <Button
              variant="tertiary"
              onClick={() => loadNext(10)}
              disabled={isLoadingNext}
              className="mt-3 mx-auto"
              icon={IconChevronDown}
            >
              {isLoadingNext && <Spinner />}
              {__("Show More")}
            </Button>
          )}
        </>
      );
}
