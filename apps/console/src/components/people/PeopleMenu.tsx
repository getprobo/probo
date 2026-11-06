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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { AvatarSkeleton } from "@probo/ui/src/v2/Avatar/AvatarSkeleton";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useDropdownContext } from "@probo/ui/src/v2/Dropdown/context";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";
import { type ReactNode, Suspense, use, useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { PeopleMenu_organization$key } from "#/__generated__/core/PeopleMenu_organization.graphql";
import type { PeopleMenuQuery } from "#/__generated__/core/PeopleMenuQuery.graphql";
import type {
  PeopleMenuRefetchQuery,
  ProfileFilter,
} from "#/__generated__/core/PeopleMenuRefetchQuery.graphql";
import { UserSelectOption } from "#/components/form/UserSelectOption";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser, type CurrentUserProfile } from "#/providers/CurrentUser";

const pageSize = 10;
const searchDebounceMs = 300;
const peopleMenuSkeletonWidths = ["w-24", "w-16", "w-28", "w-20"] as const;

function assigneeFilter(query: string | null, hideEndedContracts: boolean): ProfileFilter {
  return {
    states: ["ACTIVE", "PENDING"],
    ...(hideEndedContracts ? { contractEnded: false } : {}),
    query,
  };
}

const taskAssigneeMenuQuery = graphql`
  query PeopleMenuQuery($organizationId: ID!, $filter: ProfileFilter) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...PeopleMenu_organization @arguments(filter: $filter)
      }
    }
  }
`;

const taskAssigneeMenuFragment = graphql`
  fragment PeopleMenu_organization on Organization
  @refetchable(queryName: "PeopleMenuRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey", defaultValue: null }
    filter: { type: "ProfileFilter", defaultValue: null }
  ) {
    profiles(
      first: $first
      after: $after
      orderBy: { direction: DESC, field: CREATED_AT }
      filter: $filter
    ) @connection(key: "PeopleMenu_profiles", filters: ["filter"]) {
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

export interface PeopleOption {
  id: string;
  fullName: string;
  emailAddress: string;
  avatarUrl?: string | null;
}

export interface PeopleMenuProps {
  value: string | null;
  emptyLabel?: string;
  emptyValue: string | null;
  // Icon shown beside the empty choice. The user filter leaves this unset
  // because "All users" is not an assignee.
  emptyIcon?: ReactNode;
  // Second empty choice, used by the people filter to select tasks with no assignee.
  unassigned?: {
    label: string;
    value: string;
    icon: ReactNode;
  };
  disabled?: boolean;
  hideEndedContracts?: boolean;
  onValueChange: (value: string | null) => void;
  onSelect?: (assignee: PeopleOption | null) => void;
}

export function PeopleMenu(props: PeopleMenuProps) {
  return (
    <Suspense fallback={<PeopleMenuSkeleton disabled={props.disabled} />}>
      <PeopleMenuLoaded {...props} />
    </Suspense>
  );
}

function PeopleMenuLoaded({
  value,
  emptyLabel,
  emptyValue,
  emptyIcon,
  unassigned,
  disabled,
  hideEndedContracts = true,
  onValueChange,
  onSelect,
}: PeopleMenuProps) {
  const { t } = useTranslation();
  const { size } = useDropdownContext();
  const organizationId = useOrganizationId();
  const { profile } = use(CurrentUser);
  const [input, setInput] = useState("");
  const [graphqlQuery, setGraphqlQuery] = useState<string | null>(null);
  const [, startTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const data = useLazyLoadQuery<PeopleMenuQuery>(
    taskAssigneeMenuQuery,
    { organizationId, filter: assigneeFilter(null, hideEndedContracts) },
  );
  const {
    data: organization,
    loadNext,
    hasNext,
    isLoadingNext,
    refetch,
  } = usePaginationFragment<
    PeopleMenuRefetchQuery,
    PeopleMenu_organization$key
  >(
    taskAssigneeMenuFragment,
    data.organization.__typename === "Organization" ? data.organization : null,
  );

  useEffect(() => {
    const handle = setTimeout(() => {
      const next = input.trim();
      setGraphqlQuery(next === "" ? null : next);
    }, searchDebounceMs);
    return () => {
      clearTimeout(handle);
    };
  }, [input]);

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    startTransition(() => {
      refetch({ filter: assigneeFilter(graphqlQuery, hideEndedContracts) }, { fetchPolicy: "network-only" });
    });
  }, [graphqlQuery, hideEndedContracts, refetch]);

  if (organization == null) {
    return null;
  }

  const loaded = (organization.profiles?.edges ?? []).flatMap((edge) => {
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
  const users = graphqlQuery == null
    ? pinViewer(loaded, profile)
    : loaded;

  function select(next: string | null) {
    const resolved = next === emptyValue ? null : next;
    if (resolved === value) {
      return;
    }
    onValueChange(resolved);
    if (onSelect == null) {
      return;
    }
    if (resolved == null || (unassigned != null && resolved === unassigned.value)) {
      onSelect(null);
      return;
    }
    const selected = users.find(user => user.id === resolved);
    if (selected != null) {
      onSelect(selected);
    }
  }

  return (
    <div className="flex w-64 flex-col gap-1">
      <PeopleMenuSearch
        value={input}
        size={size}
        disabled={disabled}
        onValueChange={setInput}
      />
      <div className="flex max-h-64 flex-col gap-1 overflow-y-auto">
        {emptyLabel != null && (
          <DropdownItem
            disabled={disabled}
            onClick={() => {
              select(emptyValue);
            }}
          >
            {emptyIcon == null
              ? emptyLabel
              : (
                  <UserSelectOption
                    fullName={emptyLabel}
                    size={size}
                    fallback={emptyIcon}
                  />
                )}
          </DropdownItem>
        )}
        {unassigned != null && (
          <DropdownItem
            disabled={disabled}
            onClick={() => {
              select(unassigned.value);
            }}
          >
            <UserSelectOption
              fullName={unassigned.label}
              size={size}
              fallback={unassigned.icon}
            />
          </DropdownItem>
        )}
        {users.map(user => (
          <DropdownItem
            key={user.id}
            disabled={disabled}
            onClick={() => {
              select(user.id);
            }}
          >
            <UserSelectOption
              fullName={user.fullName}
              emailAddress={user.emailAddress}
              avatarUrl={user.avatarUrl}
              size={size}
            />
          </DropdownItem>
        ))}
        {users.length === 0 && (
          <Text size={2} color="faint" className="px-3 py-1.5">
            {t("peopleMenu.empty")}
          </Text>
        )}
      </div>
      {hasNext && (
        <Button
          variant="ghost"
          color="neutral"
          size={1}
          className="w-full"
          loading={isLoadingNext}
          disabled={disabled}
          onClick={() => {
            loadNext(pageSize);
          }}
        >
          {t("userSelect.loadMore")}
        </Button>
      )}
    </div>
  );
}

function PeopleMenuSkeleton({ disabled }: { disabled?: boolean }) {
  const { size } = useDropdownContext();

  return (
    <div className="flex w-64 flex-col gap-1" aria-busy="true">
      <PeopleMenuSearch size={size} disabled={disabled} />
      {peopleMenuSkeletonWidths.map(width => (
        <div
          key={width}
          className={size === 1
            ? "flex h-7 items-center gap-1 px-2"
            : "flex h-8 items-center gap-2 px-3"}
          aria-hidden
        >
          <AvatarSkeleton
            size={1}
            radius="full"
            className={size === 1 ? "size-4!" : undefined}
          />
          <TextSkeleton size={size} className={width} />
        </div>
      ))}
    </div>
  );
}

function PeopleMenuSearch({
  value = "",
  size,
  disabled,
  onValueChange,
}: {
  value?: string;
  size: 1 | 2;
  disabled?: boolean;
  onValueChange?: (value: string) => void;
}) {
  const { t } = useTranslation();

  return (
    <div
      className="px-1 pb-1"
      onKeyDown={(event) => {
        event.stopPropagation();
      }}
    >
      <TextField
        size={size}
        icon={<MagnifyingGlassIcon />}
        value={value}
        disabled={disabled}
        onValueChange={onValueChange}
        placeholder={t("peopleMenu.searchPlaceholder")}
        aria-label={t("peopleMenu.searchPlaceholder")}
      />
    </div>
  );
}

function pinViewer(
  users: readonly PeopleOption[],
  profile: CurrentUserProfile | null,
): PeopleOption[] {
  if (profile == null || profile.id === "") {
    return [...users];
  }

  const viewer = {
    id: profile.id,
    fullName: profile.fullName,
    emailAddress: profile.emailAddress,
    avatarUrl: profile.avatarUrl,
  };
  return [
    viewer,
    ...users.filter(user => user.id !== profile.id),
  ];
}
