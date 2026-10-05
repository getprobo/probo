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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectList } from "@probo/ui/src/v2/Select/SelectList";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Suspense, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { UserSelect_organization$key } from "#/__generated__/core/UserSelect_organization.graphql";
import type { UserSelectPaginationQuery } from "#/__generated__/core/UserSelectPaginationQuery.graphql";
import type { UserSelectProfileQuery } from "#/__generated__/core/UserSelectProfileQuery.graphql";
import type {
  ProfileFilter,
  UserSelectQuery,
} from "#/__generated__/core/UserSelectQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { UserSelectOption } from "./UserSelectOption";

const defaultPageSize = 200;

const userSelectProfileQuery = graphql`
  query UserSelectProfileQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on Profile {
        id
        fullName
        emailAddress
        avatar {
          downloadUrl
        }
      }
    }
  }
`;

const userSelectQuery = graphql`
  query UserSelectQuery(
    $organizationId: ID!
    $first: Int!
    $filter: ProfileFilter
  ) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...UserSelect_organization @arguments(first: $first, filter: $filter)
      }
    }
  }
`;

const userSelectFragment = graphql`
  fragment UserSelect_organization on Organization
  @refetchable(queryName: "UserSelectPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 200 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    filter: { type: "ProfileFilter", defaultValue: null }
  ) {
    profiles(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { direction: ASC, field: FULL_NAME }
      filter: $filter
    ) @connection(key: "UserSelect_profiles", filters: ["filter"]) {
      edges {
        node {
          id
          fullName
          emailAddress
          avatar {
            downloadUrl
          }
        }
      }
    }
  }
`;

interface UserOption {
  id: string;
  fullName: string;
  emailAddress: string;
  avatarUrl?: string | null;
}

interface UserSelectProps {
  value: string | null;
  onValueChange: (value: string | null) => void;
  emptyLabel: string;
  emptyValue?: string | null;
  ariaLabel: string;
  placeholder?: string;
  disabled?: boolean;
  size?: 1 | 2;
  pageSize?: number;
  contractEnded?: boolean;
  pinned?: UserOption | null;
}

interface UserSelectLoadedProps extends UserSelectProps {
  size: 1 | 2;
  pageSize: number;
  emptyValue: string | null;
  placeholder: string;
}

export function UserSelect({
  size = 2,
  pageSize = defaultPageSize,
  emptyValue = null,
  placeholder,
  ...props
}: UserSelectProps) {
  return (
    <Suspense fallback={<SelectSkeleton size={size} className="w-full" />}>
      <UserSelectLoaded
        {...props}
        size={size}
        pageSize={pageSize}
        emptyValue={emptyValue}
        placeholder={placeholder ?? props.emptyLabel}
      />
    </Suspense>
  );
}

function UserSelectLoaded({
  value,
  onValueChange,
  emptyLabel,
  emptyValue,
  ariaLabel,
  placeholder,
  disabled,
  size,
  pageSize,
  contractEnded,
  pinned,
}: UserSelectLoadedProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const filter = useMemo<ProfileFilter>(() => (
    contractEnded === undefined
      ? { states: ["ACTIVE", "PENDING"] }
      : { contractEnded, states: ["ACTIVE", "PENDING"] }
  ), [contractEnded]);
  const query = useLazyLoadQuery<UserSelectQuery>(
    userSelectQuery,
    { organizationId, first: pageSize, filter },
    { fetchPolicy: "network-only" },
  );
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    UserSelectPaginationQuery,
    UserSelect_organization$key
  >(
    userSelectFragment,
    query.organization as UserSelect_organization$key,
  );
  const loaded = (data.profiles?.edges ?? []).flatMap((edge) => {
    const node = edge?.node;
    if (node == null) {
      return [];
    }
    return [{
      id: node.id,
      fullName: node.fullName,
      emailAddress: node.emailAddress,
      avatarUrl: node.avatar?.downloadUrl,
    }];
  });
  const users = pinned != null && !loaded.some(user => user.id === pinned.id)
    ? [pinned, ...loaded]
    : loaded;
  const names = new Map(users.map(user => [user.id, user]));

  return (
    <Select
      value={value ?? emptyValue}
      disabled={disabled}
      onValueChange={(next: string | null) => {
        const resolved = next == null || next === emptyValue ? null : next;
        if (resolved === value) {
          return;
        }
        onValueChange(resolved);
      }}
    >
      <SelectTrigger
        size={size}
        aria-label={ariaLabel}
        placeholder={placeholder}
      >
        {(selected: string | null) => {
          if (selected == null || selected === emptyValue) {
            return emptyLabel;
          }
          return (
            <UserSelectRow
              id={selected}
              users={names}
              size={size}
            />
          );
        }}
      </SelectTrigger>
      <SelectPopup align="start">
        <SelectList>
          <SelectItem value={emptyValue}>{emptyLabel}</SelectItem>
          {value != null && value !== emptyValue && !names.has(value) && (
            <SelectItem value={value}>
              <UserSelectRow id={value} users={names} size={size} />
            </SelectItem>
          )}
          {users.map(user => (
            <SelectItem key={user.id} value={user.id}>
              <UserSelectOption
                fullName={user.fullName}
                emailAddress={user.emailAddress}
                avatarUrl={user.avatarUrl}
                size={size}
              />
            </SelectItem>
          ))}
        </SelectList>
        {hasNext && (
          <Button
            variant="ghost"
            color="neutral"
            size={1}
            className="w-full"
            loading={isLoadingNext}
            onClick={() => {
              loadNext(pageSize);
            }}
          >
            {t("userSelect.loadMore")}
          </Button>
        )}
      </SelectPopup>
    </Select>
  );
}

function UserSelectRow({
  id,
  users,
  size,
}: {
  id: string;
  users: Map<string, UserOption>;
  size: 1 | 2;
}) {
  const user = users.get(id);
  if (user != null) {
    return (
      <UserSelectOption
        fullName={user.fullName}
        emailAddress={user.emailAddress}
        avatarUrl={user.avatarUrl}
        size={size}
      />
    );
  }

  return (
    <Suspense fallback={id}>
      <UserSelectProfile id={id} size={size} />
    </Suspense>
  );
}

function UserSelectProfile({ id, size }: { id: string; size: 1 | 2 }) {
  const data = useLazyLoadQuery<UserSelectProfileQuery>(
    userSelectProfileQuery,
    { id },
    { fetchPolicy: "store-or-network" },
  );
  const profile = data.node?.__typename === "Profile" ? data.node : null;
  if (profile == null) {
    return id;
  }

  return (
    <UserSelectOption
      fullName={profile.fullName}
      emailAddress={profile.emailAddress}
      avatarUrl={profile.avatar?.downloadUrl}
      size={size}
    />
  );
}
