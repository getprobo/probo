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

import { getAssignableRoles } from "@probo/helpers";
import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import { use } from "react";
import { useTranslation } from "react-i18next";
import { ConnectionHandler, graphql, usePaginationFragment } from "react-relay";

import type { PeopleListFragment$key } from "#/__generated__/iam/PeopleListFragment.graphql";
import type { PeopleListFragment_RefetchQuery } from "#/__generated__/iam/PeopleListFragment_RefetchQuery.graphql";
import { type Order, SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import { PeopleListItem } from "./PeopleListItem";

const fragment = graphql`
  fragment PeopleListFragment on Organization
  @refetchable(queryName: "PeopleListFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: {
      type: "ProfileOrder"
      defaultValue: { direction: ASC, field: FULL_NAME }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    profiles(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "PeopleListFragment_profiles", filters: ["orderBy"]) @required(action: THROW) {
      __id
      totalCount
      edges @required(action: THROW) {
        node {
          id
          ...PeopleListItemFragment
        }
      }
    }
  }
`;

export function PeopleList(props: {
  fKey: PeopleListFragment$key;
  onConnectionIdChange: (connectionId: string) => void;
}) {
  const { fKey, onConnectionIdChange } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const { role } = use(CurrentUser);
  const canManageRoles = getAssignableRoles(role).length > 0;

  const peoplePagination = usePaginationFragment<
    PeopleListFragment_RefetchQuery,
    PeopleListFragment$key
  >(fragment, fKey);

  const refetchPeople = () => {
    peoplePagination.refetch({}, { fetchPolicy: "network-only" });
  };

  const handleOrderChange = (order: Order) => {
    onConnectionIdChange(
      ConnectionHandler.getConnectionID(
        organizationId,
        "PeopleListFragment_profiles",
        { orderBy: order },
      ),
    );
  };

  return (
    <SortableTable
      {...peoplePagination}
      refetch={
        peoplePagination.refetch as ComponentProps<
          typeof SortableTable
        >["refetch"]
      }
      pageSize={20}
    >
      <Thead>
        <Tr>
          <SortableTh field="FULL_NAME" onOrderChange={handleOrderChange}>{t("peopleList.columns.name")}</SortableTh>
          <SortableTh field="STATE">{t("peopleList.columns.status")}</SortableTh>
          <SortableTh field="EMAIL_ADDRESS" onOrderChange={handleOrderChange}>{t("peopleList.columns.email")}</SortableTh>
          {canManageRoles && <Th>{t("peopleList.columns.role")}</Th>}
          <SortableTh field="CREATED_AT" onOrderChange={handleOrderChange}>{t("peopleList.columns.createdOn")}</SortableTh>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {peoplePagination.data.profiles.totalCount === 0
          ? (
              <Tr>
                <Td colSpan={7} className="text-center text-txt-secondary">
                  {t("peopleList.empty")}
                </Td>
              </Tr>
            )
          : (
              peoplePagination.data.profiles.edges.map(({ node: profile }) => (
                <PeopleListItem
                  connectionId={peoplePagination.data.profiles.__id}
                  key={profile.id}
                  fKey={profile}
                  onRefetch={refetchPeople}
                />
              ))
            )}
      </Tbody>
    </SortableTable>
  );
}
