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
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { useParams } from "react-router";

import type { CompliancePortalAccessListFragment$key } from "#/__generated__/core/CompliancePortalAccessListFragment.graphql";
import type { CompliancePortalAccessListQuery } from "#/__generated__/core/CompliancePortalAccessListQuery.graphql";
import type { CompliancePortalAccessListRootQuery } from "#/__generated__/core/CompliancePortalAccessListRootQuery.graphql";

import { CompliancePortalAccessListItem } from "./CompliancePortalAccessListItem";

const accessListQuery = graphql`
  query CompliancePortalAccessListRootQuery($compliancePortalId: ID!) {
    node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalAccessListFragment
      }
    }
  }
`;

const fragment = graphql`
  fragment CompliancePortalAccessListFragment on CompliancePortal
  @argumentDefinitions(
    first: { type: Int, defaultValue: 10 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: CompliancePortalAccessOrder, defaultValue: { field: CREATED_AT, direction: DESC } }
  )
  @refetchable(queryName: "CompliancePortalAccessListQuery") {
    accesses(
      first: $first
      after: $after
      orderBy: $order
    ) @connection(key: "CompliancePortalAccessList_accesses" filters: ["orderBy"]) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...CompliancePortalAccessListItemFragment
        }
      }
    }
  }
`;

export function CompliancePortalAccessList() {
  const { t } = useTranslation("organizations/compliance-portals");
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const { node } = useLazyLoadQuery<CompliancePortalAccessListRootQuery>(
    accessListQuery,
    { compliancePortalId: compliancePortalId ?? "" },
  );
  const portalKey = node?.__typename === "CompliancePortal" ? node : null;
  const {
    data,
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment<CompliancePortalAccessListQuery, CompliancePortalAccessListFragment$key>(
    fragment,
    portalKey,
  );
  if (compliancePortalId == null || data == null) {
    throw new Error("invalid type for node");
  }

  const { accesses } = data;

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
                <CompliancePortalAccessListItem
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
