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

import { getAssignableRoles, getMembershipRoles, peopleRoles } from "@probo/helpers";
import {
  IconMagnifyingGlass,
  Input,
  Option,
  Select,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import type { ComponentProps } from "react";
import { use, useCallback, useEffect, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";
import { useDebounceCallback } from "usehooks-ts";

import type { PeopleListFragment$key } from "#/__generated__/iam/PeopleListFragment.graphql";
import type {
  MembershipRole,
  PeopleListFragment_RefetchQuery,
  ProfileOrderField,
  ProfileState,
} from "#/__generated__/iam/PeopleListFragment_RefetchQuery.graphql";
import { type Order, SortableTable, SortableTh } from "#/components/SortableTable";
import { CurrentUser } from "#/providers/CurrentUser";

import { PeopleListItem } from "./PeopleListItem";

const PAGE_SIZE = 100;
const SEARCH_DEBOUNCE_MS = 300;

type PeopleKind = (typeof peopleRoles)[number];

const fragment = graphql`
  fragment PeopleListFragment on Organization
  @refetchable(queryName: "PeopleListFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    order: {
      type: "ProfileOrder"
      defaultValue: { direction: ASC, field: FULL_NAME }
    }
    filter: { type: "ProfileFilter", defaultValue: null }
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
      filter: $filter
    ) @connection(key: "PeopleListFragment_profiles", filters: ["orderBy", "filter"]) @required(action: THROW) {
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

type PeopleFilter = {
  query: string | null;
  state: ProfileState | null;
  role: MembershipRole | null;
  kind: string | null;
};

export function PeopleList(props: {
  fKey: PeopleListFragment$key;
  onConnectionIdChange: (connectionId: string) => void;
}) {
  const { fKey, onConnectionIdChange } = props;

  const { t } = useTranslation();
  const { role } = use(CurrentUser);
  const canManageRoles = getAssignableRoles(role).length > 0;

  const [queryFilter, setQueryFilter] = useState<string | null>(null);
  const [stateFilter, setStateFilter] = useState<ProfileState | null>(null);
  const [roleFilter, setRoleFilter] = useState<MembershipRole | null>(null);
  const [kindFilter, setKindFilter] = useState<string | null>(null);
  const [order, setOrder] = useState<Order>({
    direction: "ASC",
    field: "FULL_NAME",
  });
  const [isPending, startTransition] = useTransition();

  const peoplePagination = usePaginationFragment<
    PeopleListFragment_RefetchQuery,
    PeopleListFragment$key
  >(fragment, fKey);

  const connectionId = peoplePagination.data.profiles.__id;

  useEffect(() => {
    onConnectionIdChange(connectionId);
  }, [connectionId, onConnectionIdChange]);

  const currentFilter = (overrides: Partial<PeopleFilter> = {}): PeopleFilter => ({
    query: queryFilter,
    state: stateFilter,
    role: roleFilter,
    kind: kindFilter,
    ...overrides,
  });

  const connectionFilter = (filter: PeopleFilter) => ({
    query: filter.query,
    state: filter.state,
    role: filter.role,
    kind: filter.kind,
    contractEnded: null,
  });

  const refetchPeople = (overrides: Partial<PeopleFilter> = {}, nextOrder: Order = order) => {
    const filter = currentFilter(overrides);
    startTransition(() => {
      peoplePagination.refetch(
        {
          order: {
            direction: nextOrder.direction,
            field: nextOrder.field as ProfileOrderField,
          },
          filter: connectionFilter(filter),
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const debouncedRefetchQuery = useDebounceCallback(
    useCallback(
      (value: string) => {
        const newQuery = value === "" ? null : value;
        startTransition(() => {
          peoplePagination.refetch(
            {
              order: {
                direction: order.direction,
                field: order.field as ProfileOrderField,
              },
              filter: {
                query: newQuery,
                state: stateFilter,
                role: roleFilter,
                kind: kindFilter,
                contractEnded: null,
              },
            },
            { fetchPolicy: "network-only" },
          );
        });
      },
      [peoplePagination, order, stateFilter, roleFilter, kindFilter],
    ),
    SEARCH_DEBOUNCE_MS,
  );

  const handleQueryFilterChange = (value: string) => {
    setQueryFilter(value === "" ? null : value);
    debouncedRefetchQuery(value);
  };

  const handleStateFilterChange = (value: string) => {
    const newState = value === "ALL" ? null : (value as ProfileState);
    setStateFilter(newState);
    refetchPeople({ state: newState });
  };

  const handleRoleFilterChange = (value: string) => {
    const newRole = value === "ALL" ? null : (value as MembershipRole);
    setRoleFilter(newRole);
    refetchPeople({ role: newRole });
  };

  const handleKindFilterChange = (value: string) => {
    const newKind = value === "ALL" ? null : value;
    setKindFilter(newKind);
    refetchPeople({ kind: newKind });
  };

  const handleOrderChange = (nextOrder: Order) => {
    setOrder(nextOrder);
  };

  const refetchWithFilters: ComponentProps<typeof SortableTable>["refetch"] = ({ order: nextOrder }) => {
    setOrder(nextOrder);
    refetchPeople({}, nextOrder);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={t("peopleList.searchPlaceholder")}
          value={queryFilter ?? ""}
          onValueChange={handleQueryFilterChange}
        />
        <Select
          value={stateFilter ?? "ALL"}
          onValueChange={handleStateFilterChange}
        >
          <Option value="ALL">{t("peopleList.filters.allStatuses")}</Option>
          <Option value="ACTIVE">{t("peopleList.filters.active")}</Option>
          <Option value="INACTIVE">{t("peopleList.filters.inactive")}</Option>
        </Select>
        <Select
          value={roleFilter ?? "ALL"}
          onValueChange={handleRoleFilterChange}
        >
          <Option value="ALL">{t("peopleList.filters.allRoles")}</Option>
          {getMembershipRoles(t).map(({ value, label }) => (
            <Option key={value} value={value}>
              {label}
            </Option>
          ))}
        </Select>
        <Select
          value={kindFilter ?? "ALL"}
          onValueChange={handleKindFilterChange}
        >
          <Option value="ALL">{t("peopleList.filters.allTypes")}</Option>
          {peopleRoles.map((kind: PeopleKind) => (
            <Option key={kind} value={kind}>
              {t(`personForm.kinds.${kind}`)}
            </Option>
          ))}
        </Select>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        <SortableTable
          {...peoplePagination}
          refetch={refetchWithFilters}
          pageSize={PAGE_SIZE}
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
                      connectionId={connectionId}
                      key={profile.id}
                      fKey={profile}
                      onRefetch={() => refetchPeople()}
                    />
                  ))
                )}
          </Tbody>
        </SortableTable>
      </div>
    </div>
  );
}
