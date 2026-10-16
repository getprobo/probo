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
import { Suspense, use, useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { TaskAssigneeMenu_organization$key } from "#/__generated__/core/TaskAssigneeMenu_organization.graphql";
import type { TaskAssigneeMenuQuery } from "#/__generated__/core/TaskAssigneeMenuQuery.graphql";
import type {
  ProfileFilter,
  TaskAssigneeMenuRefetchQuery,
} from "#/__generated__/core/TaskAssigneeMenuRefetchQuery.graphql";
import { UserSelectOption } from "#/components/form/UserSelectOption";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser, type CurrentUserProfile } from "#/providers/CurrentUser";

const pageSize = 10;
const searchDebounceMs = 300;
const assigneeMenuSkeletonWidths = ["w-24", "w-16", "w-28", "w-20"] as const;

function assigneeFilter(query: string | null, hideEndedContracts: boolean): ProfileFilter {
  return {
    states: ["ACTIVE", "PENDING"],
    ...(hideEndedContracts ? { contractEnded: false } : {}),
    query,
  };
}

const taskAssigneeMenuQuery = graphql`
  query TaskAssigneeMenuQuery($organizationId: ID!, $filter: ProfileFilter) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...TaskAssigneeMenu_organization @arguments(filter: $filter)
      }
    }
  }
`;

const taskAssigneeMenuFragment = graphql`
  fragment TaskAssigneeMenu_organization on Organization
  @refetchable(queryName: "TaskAssigneeMenuRefetchQuery")
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
    ) @connection(key: "TaskAssigneeMenu_profiles", filters: ["filter"]) {
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

interface AssigneeOption {
  id: string;
  fullName: string;
  emailAddress: string;
  avatarUrl?: string | null;
}

export interface TaskAssigneeMenuProps {
  value: string | null;
  emptyLabel: string;
  emptyValue: string | null;
  disabled?: boolean;
  hideEndedContracts?: boolean;
  onValueChange: (value: string | null) => void;
}

export function TaskAssigneeMenu(props: TaskAssigneeMenuProps) {
  return (
    <Suspense fallback={<AssigneeMenuSkeleton disabled={props.disabled} />}>
      <TaskAssigneeMenuLoaded {...props} />
    </Suspense>
  );
}

function TaskAssigneeMenuLoaded({
  value,
  emptyLabel,
  emptyValue,
  disabled,
  hideEndedContracts = true,
  onValueChange,
}: TaskAssigneeMenuProps) {
  const { t } = useTranslation("organizations/tasks");
  const { t: tCommon } = useTranslation();
  const { size } = useDropdownContext();
  const organizationId = useOrganizationId();
  const { profile } = use(CurrentUser);
  const [input, setInput] = useState("");
  const [graphqlQuery, setGraphqlQuery] = useState<string | null>(null);
  const [, startTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const data = useLazyLoadQuery<TaskAssigneeMenuQuery>(
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
    TaskAssigneeMenuRefetchQuery,
    TaskAssigneeMenu_organization$key
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
      refetch({ filter: assigneeFilter(graphqlQuery, hideEndedContracts) }, { fetchPolicy: "store-or-network" });
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
  }

  return (
    <div className="flex w-64 flex-col gap-1">
      <AssigneeMenuSearch
        value={input}
        size={size}
        disabled={disabled}
        onValueChange={setInput}
      />
      <DropdownItem
        disabled={disabled}
        onClick={() => {
          select(emptyValue);
        }}
      >
        {emptyLabel}
      </DropdownItem>
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
          {t("detailsPage.assigneeMenu.empty")}
        </Text>
      )}
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
          {tCommon("userSelect.loadMore")}
        </Button>
      )}
    </div>
  );
}

function AssigneeMenuSkeleton({ disabled }: { disabled?: boolean }) {
  const { size } = useDropdownContext();

  return (
    <div className="flex w-64 flex-col gap-1" aria-busy="true">
      <AssigneeMenuSearch size={size} disabled={disabled} />
      {assigneeMenuSkeletonWidths.map(width => (
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

function AssigneeMenuSearch({
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
  const { t } = useTranslation("organizations/tasks");

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
        placeholder={t("detailsPage.assigneeMenu.searchPlaceholder")}
        aria-label={t("detailsPage.assigneeMenu.searchPlaceholder")}
      />
    </div>
  );
}

function pinViewer(
  users: readonly AssigneeOption[],
  profile: CurrentUserProfile | null,
): AssigneeOption[] {
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
