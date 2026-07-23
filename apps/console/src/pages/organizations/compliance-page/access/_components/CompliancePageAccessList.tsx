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

import { Button, IconChevronDown, Spinner, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAccessListFragment$key } from "#/__generated__/core/CompliancePageAccessListFragment.graphql";
import type { CompliancePageAccessListQuery } from "#/__generated__/core/CompliancePageAccessListQuery.graphql";

import { CompliancePageAccessListItem } from "./CompliancePageAccessListItem";

const fragment = graphql`
  fragment CompliancePageAccessListFragment on CompliancePortal
  @argumentDefinitions(
    first: { type: Int, defaultValue: 10 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: CompliancePortalAccessOrder, defaultValue: { field: CREATED_AT, direction: DESC } }
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

  const { t } = useTranslation("organizations/compliance-page");

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
                {t("accessList.empty")}
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
                <Th>{t("accessList.columns.name")}</Th>
                <Th>{t("accessList.columns.email")}</Th>
                <Th>{t("accessList.columns.date")}</Th>
                <Th className="text-center">
                  {t("accessList.columns.access")}
                </Th>
                <Th className="text-center">
                  {t("accessList.columns.requests")}
                </Th>
                <Th className="text-center">{t("accessList.columns.nda")}</Th>
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
              {t("accessList.actions.showMore")}
            </Button>
          )}
        </>
      );
}
