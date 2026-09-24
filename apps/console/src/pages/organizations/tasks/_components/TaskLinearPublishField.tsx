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
import { Combobox } from "@probo/ui/src/v2/Combobox/Combobox";
import { ComboboxEmpty } from "@probo/ui/src/v2/Combobox/ComboboxEmpty";
import { ComboboxInput } from "@probo/ui/src/v2/Combobox/ComboboxInput";
import { ComboboxInputGroup } from "@probo/ui/src/v2/Combobox/ComboboxInputGroup";
import { ComboboxItem } from "@probo/ui/src/v2/Combobox/ComboboxItem";
import { ComboboxList } from "@probo/ui/src/v2/Combobox/ComboboxList";
import { ComboboxPopup } from "@probo/ui/src/v2/Combobox/ComboboxPopup";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, graphql, useFragment, useRelayEnvironment } from "react-relay";
import type { GraphQLTaggedNode } from "relay-runtime";

import type { TaskLinearPublishField_task$key } from "#/__generated__/core/TaskLinearPublishField_task.graphql";
import type { TaskLinearPublishFieldIssueQuery } from "#/__generated__/core/TaskLinearPublishFieldIssueQuery.graphql";
import type { TaskLinearPublishFieldLinkMutation } from "#/__generated__/core/TaskLinearPublishFieldLinkMutation.graphql";
import type { TaskLinearPublishFieldMutation } from "#/__generated__/core/TaskLinearPublishFieldMutation.graphql";
import type { TaskLinearPublishFieldTeamQuery } from "#/__generated__/core/TaskLinearPublishFieldTeamQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import type { TaskPriority, TaskState } from "../_lib/taskState";
import { taskLinearPublishField } from "../variants";

const PAGE_SIZE = 20;
const NEW_ISSUE_ID = "__new__";

interface LinearIssueOption {
  id: string;
  identifier: string;
  title: string;
  content: string | null;
  state: TaskState | null;
  priority: TaskPriority | null;
}

export interface LinearIssueDraft {
  title: string;
  content: string | null;
  state: TaskState | null;
  priority: TaskPriority | null;
}

function linearIssueOption(node: {
  id: string;
  identifier: string;
  title: string;
  content?: string | null;
  state?: TaskState | null;
  priority?: TaskPriority | null;
}): LinearIssueOption {
  return {
    id: node.id,
    identifier: node.identifier,
    title: node.title,
    content: node.content ?? null,
    state: node.state ?? null,
    priority: node.priority ?? null,
  };
}

const taskLinearPublishFieldFragment = graphql`
  fragment TaskLinearPublishField_task on Task {
    id
  }
`;

const teamPageQuery = graphql`
  query TaskLinearPublishFieldTeamQuery(
    $organizationId: ID!
    $first: Int!
    $after: String
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        linearTeams(first: $first, after: $after, query: $query) {
          edges {
            node {
              id
              name
              key
            }
          }
          pageInfo {
            hasNextPage
            endCursor
          }
        }
      }
    }
  }
`;

const issuePageQuery = graphql`
  query TaskLinearPublishFieldIssueQuery(
    $organizationId: ID!
    $teamId: String!
    $first: Int!
    $after: String
    $query: String
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        linearIssues(teamId: $teamId, first: $first, after: $after, query: $query) {
          edges {
            node {
              id
              identifier
              title
              content
              state
              priority
            }
          }
          pageInfo {
            hasNextPage
            endCursor
          }
        }
      }
    }
  }
`;

const publishMutation = graphql`
  mutation TaskLinearPublishFieldMutation($input: PublishTaskToLinearInput!) {
    publishTaskToLinear(input: $input) {
      task {
        ...TaskLinearField_task
      }
    }
  }
`;

const linkMutation = graphql`
  mutation TaskLinearPublishFieldLinkMutation($input: LinkTaskToLinearInput!) {
    linkTaskToLinear(input: $input) {
      task {
        ...TaskLinearField_task
        ...TaskDetailsPage_task
        ...TasksCard_TaskRowFragment
      }
    }
  }
`;

interface TaskLinearPublishFieldProps {
  taskKey: TaskLinearPublishField_task$key;
}

interface PageResult<T> {
  items: T[];
  hasNextPage: boolean;
  endCursor: string | null;
}

function useDebounced(value: string, delay: number) {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timeout = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timeout);
  }, [value, delay]);

  return debounced;
}

export function TaskLinearPublishField({ taskKey }: TaskLinearPublishFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const task = useFragment(taskLinearPublishFieldFragment, taskKey);
  const [teamId, setTeamId] = useState<string | null>(null);
  const [teamLabel, setTeamLabel] = useState<string | null>(null);
  const [issueId, setIssueId] = useState(NEW_ISSUE_ID);
  const [issueLabel, setIssueLabel] = useState<string | null>(null);
  const [confirmLink, setConfirmLink] = useState(false);
  const [publish, isPublishing] = useMutation<TaskLinearPublishFieldMutation>(
    publishMutation,
    {
      successMessage: t("detailsPage.linear.published"),
      errorToast: t("detailsPage.linear.errors.publish"),
    },
  );
  const [link, isLinking] = useMutation<TaskLinearPublishFieldLinkMutation>(
    linkMutation,
    {
      successMessage: t("detailsPage.linear.linked"),
      errorToast: t("detailsPage.linear.errors.link"),
    },
  );
  const { root } = taskLinearPublishField();
  const isSaving = isPublishing || isLinking;
  const linking = issueId !== NEW_ISSUE_ID;
  const newIssueLabel = t("detailsPage.linear.newIssue");

  return (
    <div className={root()}>
      <RemoteCombobox<TaskLinearPublishFieldTeamQuery, { id: string; name: string; key: string }>
        query={teamPageQuery}
        variables={(after, query) => ({
          organizationId,
          first: PAGE_SIZE,
          after,
          query: query || null,
        })}
        readPage={(data) => {
          const organization = data.node?.__typename === "Organization" ? data.node : null;
          const page = organization?.linearTeams;
          return {
            items: page?.edges.map(edge => edge.node) ?? [],
            hasNextPage: page?.pageInfo.hasNextPage ?? false,
            endCursor: page?.pageInfo.endCursor ?? null,
          };
        }}
        reloadKey={organizationId}
        selectedId={teamId}
        selectedLabel={teamLabel ?? t("detailsPage.linear.chooseTeam")}
        placeholder={t("detailsPage.linear.searchTeam")}
        emptyLabel={t("detailsPage.linear.noResults")}
        disabled={isSaving}
        onSelect={(team) => {
          setTeamId(team.id);
          setTeamLabel(`${team.name} (${team.key})`);
          setIssueId(NEW_ISSUE_ID);
          setIssueLabel(null);
        }}
        itemId={team => team.id}
        itemLabel={team => `${team.name} (${team.key})`}
      />
      <RemoteCombobox<TaskLinearPublishFieldIssueQuery, LinearIssueOption>
        query={issuePageQuery}
        variables={(after, query) => ({
          organizationId,
          teamId: teamId ?? "",
          first: PAGE_SIZE,
          after,
          query: query || null,
        })}
        readPage={(data) => {
          const organization = data.node?.__typename === "Organization" ? data.node : null;
          const page = organization?.linearIssues;
          return {
            items: page?.edges.map(edge => linearIssueOption(edge.node)) ?? [],
            hasNextPage: page?.pageInfo.hasNextPage ?? false,
            endCursor: page?.pageInfo.endCursor ?? null,
          };
        }}
        reloadKey={teamId ?? ""}
        enabled={teamId != null}
        selectedId={teamId == null ? null : issueId}
        selectedLabel={
          teamId == null
            ? t("detailsPage.linear.chooseIssue")
            : issueId === NEW_ISSUE_ID
              ? newIssueLabel
              : (issueLabel ?? newIssueLabel)
        }
        placeholder={t("detailsPage.linear.searchIssue")}
        emptyLabel={t("detailsPage.linear.noResults")}
        disabled={isSaving || teamId == null}
        pinned={{
          id: NEW_ISSUE_ID,
          label: newIssueLabel,
        }}
        onSelect={(issue) => {
          setIssueId(issue.id);
          setIssueLabel(`${issue.identifier} ${issue.title}`);
        }}
        onPinnedSelect={() => {
          setIssueId(NEW_ISSUE_ID);
          setIssueLabel(null);
        }}
        itemId={issue => issue.id}
        itemLabel={issue => `${issue.identifier} ${issue.title}`}
      />
      <Button
        size={1}
        disabled={teamId == null || isSaving || (linking && issueId === NEW_ISSUE_ID)}
        onClick={() => {
          if (teamId == null) {
            return;
          }

          if (linking) {
            setConfirmLink(true);
            return;
          }

          void publish({
            variables: {
              input: {
                taskId: task.id,
                teamId,
              },
            },
          });
        }}
      >
        {linking ? t("detailsPage.linear.link") : t("detailsPage.linear.publish")}
      </Button>
      <Dialog open={confirmLink} onOpenChange={setConfirmLink}>
        <DialogPopup>
          <DialogHeader>
            <DialogTitle>{t("detailsPage.linear.confirmLink.title")}</DialogTitle>
            <DialogDescription>
              {t("detailsPage.linear.confirmLink.description")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral" disabled={isSaving}>
                  {t("detailsPage.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="button"
              variant="solid"
              color="neutral"
              highContrast
              loading={isLinking}
              disabled={teamId == null}
              onClick={() => {
                if (teamId == null) {
                  return;
                }

                void link({
                  variables: {
                    input: {
                      taskId: task.id,
                      teamId,
                      issueId,
                    },
                  },
                }).then(
                  () => {
                    setConfirmLink(false);
                  },
                  () => {
                    // The Linear toast already explains the failure.
                  },
                );
              }}
            >
              {t("detailsPage.linear.link")}
            </Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>
    </div>
  );
}

export interface LinearDraft {
  teamId: string | null;
  issueId: string | null;
}

export function TaskLinearDraftField({
  disabled,
  onChange,
  onIssue,
}: {
  disabled?: boolean;
  onChange: (selection: LinearDraft) => void;
  onIssue?: (issue: LinearIssueDraft) => void;
}) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const [teamId, setTeamId] = useState<string | null>(null);
  const [teamLabel, setTeamLabel] = useState<string | null>(null);
  const [issueId, setIssueId] = useState(NEW_ISSUE_ID);
  const [issueLabel, setIssueLabel] = useState<string | null>(null);
  const { root } = taskLinearPublishField();
  const newIssueLabel = t("detailsPage.linear.newIssue");

  function emit(nextTeamId: string | null, nextIssueId: string) {
    onChange({
      teamId: nextTeamId,
      issueId: nextIssueId === NEW_ISSUE_ID ? null : nextIssueId,
    });
  }

  return (
    <div className={root()}>
      <RemoteCombobox<TaskLinearPublishFieldTeamQuery, { id: string; name: string; key: string }>
        query={teamPageQuery}
        variables={(after, query) => ({
          organizationId,
          first: PAGE_SIZE,
          after,
          query: query || null,
        })}
        readPage={(data) => {
          const organization = data.node?.__typename === "Organization" ? data.node : null;
          const page = organization?.linearTeams;
          return {
            items: page?.edges.map(edge => edge.node) ?? [],
            hasNextPage: page?.pageInfo.hasNextPage ?? false,
            endCursor: page?.pageInfo.endCursor ?? null,
          };
        }}
        reloadKey={organizationId}
        selectedId={teamId}
        selectedLabel={teamLabel ?? t("detailsPage.linear.chooseTeam")}
        placeholder={t("detailsPage.linear.searchTeam")}
        emptyLabel={t("detailsPage.linear.noResults")}
        disabled={disabled}
        onSelect={(team) => {
          setTeamId(team.id);
          setTeamLabel(`${team.name} (${team.key})`);
          setIssueId(NEW_ISSUE_ID);
          setIssueLabel(null);
          emit(team.id, NEW_ISSUE_ID);
        }}
        itemId={team => team.id}
        itemLabel={team => `${team.name} (${team.key})`}
      />
      <RemoteCombobox<TaskLinearPublishFieldIssueQuery, LinearIssueOption>
        query={issuePageQuery}
        variables={(after, query) => ({
          organizationId,
          teamId: teamId ?? "",
          first: PAGE_SIZE,
          after,
          query: query || null,
        })}
        readPage={(data) => {
          const organization = data.node?.__typename === "Organization" ? data.node : null;
          const page = organization?.linearIssues;
          return {
            items: page?.edges.map(edge => linearIssueOption(edge.node)) ?? [],
            hasNextPage: page?.pageInfo.hasNextPage ?? false,
            endCursor: page?.pageInfo.endCursor ?? null,
          };
        }}
        reloadKey={teamId ?? ""}
        enabled={teamId != null}
        selectedId={teamId == null ? null : issueId}
        selectedLabel={
          teamId == null
            ? t("detailsPage.linear.chooseIssue")
            : issueId === NEW_ISSUE_ID
              ? newIssueLabel
              : (issueLabel ?? newIssueLabel)
        }
        placeholder={t("detailsPage.linear.searchIssue")}
        emptyLabel={t("detailsPage.linear.noResults")}
        disabled={disabled || teamId == null}
        pinned={{
          id: NEW_ISSUE_ID,
          label: newIssueLabel,
        }}
        onSelect={(issue) => {
          setIssueId(issue.id);
          setIssueLabel(`${issue.identifier} ${issue.title}`);
          emit(teamId, issue.id);
          onIssue?.({
            title: issue.title,
            content: issue.content,
            state: issue.state,
            priority: issue.priority,
          });
        }}
        onPinnedSelect={() => {
          setIssueId(NEW_ISSUE_ID);
          setIssueLabel(null);
          emit(teamId, NEW_ISSUE_ID);
        }}
        itemId={issue => issue.id}
        itemLabel={issue => `${issue.identifier} ${issue.title}`}
      />
    </div>
  );
}

interface PinnedOption {
  id: string;
  label: string;
}

interface RemoteComboboxProps<TQuery extends { response: unknown; variables: object }, TItem> {
  query: GraphQLTaggedNode;
  variables: (after: string | null, query: string) => TQuery["variables"];
  readPage: (data: TQuery["response"]) => PageResult<TItem>;
  reloadKey: string;
  selectedId: string | null;
  selectedLabel: string;
  placeholder: string;
  emptyLabel: string;
  enabled?: boolean;
  disabled?: boolean;
  pinned?: PinnedOption;
  onSelect: (item: TItem) => void;
  onPinnedSelect?: () => void;
  onFirstItem?: (item: TItem) => void;
  itemId: (item: TItem) => string;
  itemLabel: (item: TItem) => string;
}

function RemoteCombobox<TQuery extends { response: unknown; variables: object }, TItem>({
  query,
  variables,
  readPage,
  reloadKey,
  selectedId,
  selectedLabel,
  placeholder,
  emptyLabel,
  enabled = true,
  disabled,
  pinned,
  onSelect,
  onPinnedSelect,
  onFirstItem,
  itemId,
  itemLabel,
}: RemoteComboboxProps<TQuery, TItem>) {
  const { t } = useTranslation("organizations/tasks");
  const environment = useRelayEnvironment();
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebounced(search, 300);
  const [items, setItems] = useState<TItem[]>([]);
  const [hasNext, setHasNext] = useState(false);
  const [endCursor, setEndCursor] = useState<string | null>(null);
  const [request, setRequest] = useState({ after: null as string | null, append: false, id: 0 });
  const [settledId, setSettledId] = useState<number | null>(null);
  const [failedId, setFailedId] = useState<number | null>(null);
  const variablesRef = useRef(variables);
  const readPageRef = useRef(readPage);
  const onFirstItemRef = useRef(onFirstItem);
  const [scope, setScope] = useState(reloadKey);
  const [searchScope, setSearchScope] = useState(debouncedSearch);
  const loading = enabled && settledId !== request.id;
  const failed = failedId === request.id;

  if (scope !== reloadKey) {
    setScope(reloadKey);
    setSearch("");
    setRequest({ after: null, append: false, id: request.id + 1 });
    setItems([]);
    setHasNext(false);
    setEndCursor(null);
  } else if (searchScope !== debouncedSearch) {
    setSearchScope(debouncedSearch);
    setRequest({ after: null, append: false, id: request.id + 1 });
    setItems([]);
    setHasNext(false);
    setEndCursor(null);
  }

  useEffect(() => {
    variablesRef.current = variables;
    readPageRef.current = readPage;
    onFirstItemRef.current = onFirstItem;
  });

  useEffect(() => {
    if (!enabled) {
      return;
    }

    let active = true;
    const { after, append, id } = request;
    const subscription = fetchQuery<TQuery>(
      environment,
      query,
      variablesRef.current(after, debouncedSearch),
      { fetchPolicy: "network-only" },
    ).subscribe({
      next: (data) => {
        if (!active) {
          return;
        }

        const page = readPageRef.current(data);
        setItems(current => (append ? [...current, ...page.items] : page.items));
        setHasNext(page.hasNextPage && page.endCursor != null);
        setEndCursor(page.endCursor);
        setFailedId(null);
        setSettledId(id);
        if (!append && debouncedSearch === "" && page.items[0] != null) {
          onFirstItemRef.current?.(page.items[0]);
        }
      },
      error: () => {
        if (!active) {
          return;
        }

        setFailedId(id);
        setSettledId(id);
      },
    });

    return () => {
      active = false;
      subscription.unsubscribe();
    };
  }, [debouncedSearch, enabled, environment, query, reloadKey, request]);

  const showEmpty = !loading && !failed && items.length === 0 && pinned == null;

  return (
    <Combobox
      value={selectedId}
      onValueChange={(value) => {
        if (typeof value !== "string") {
          return;
        }
        if (pinned != null && value === pinned.id) {
          onPinnedSelect?.();
          return;
        }
        const item = items.find(candidate => itemId(candidate) === value);
        if (item != null) {
          onSelect(item);
        }
      }}
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (next) {
          setSearch("");
        }
      }}
      inputValue={open ? search : (selectedId == null ? "" : selectedLabel)}
      onInputValueChange={(value, details) => {
        if (!open) {
          return;
        }
        if (details.reason !== "input-change") {
          return;
        }
        setSearch(value);
      }}
      disabled={disabled}
      filter={null}
    >
      <ComboboxInputGroup>
        <ComboboxInput
          placeholder={selectedId == null && !open ? selectedLabel : placeholder}
          aria-label={selectedLabel}
          disabled={disabled}
        />
      </ComboboxInputGroup>
      <ComboboxPopup>
        {(pinned != null || items.length > 0) && (
          <ComboboxList>
            {pinned != null && (
              <ComboboxItem
                key={pinned.id}
                value={pinned.id}
                onClick={onPinnedSelect}
              >
                {pinned.label}
              </ComboboxItem>
            )}
            {items.map(item => (
              <ComboboxItem
                key={itemId(item)}
                value={itemId(item)}
                onClick={() => onSelect(item)}
              >
                {itemLabel(item)}
              </ComboboxItem>
            ))}
          </ComboboxList>
        )}
        {loading && items.length === 0 && <Spinner size={1} />}
        {failed && (
          <Text size={2} color="faint" className="px-3 py-2">
            {t("detailsPage.linear.errors.load")}
          </Text>
        )}
        {showEmpty && <ComboboxEmpty>{emptyLabel}</ComboboxEmpty>}
        {hasNext && (
          <div
            onMouseDown={(event) => {
              event.preventDefault();
            }}
          >
            <Button
              variant="ghost"
              color="neutral"
              size={1}
              className="w-full justify-center"
              loading={loading}
              onClick={() => {
                if (endCursor == null || loading) {
                  return;
                }

                setRequest(current => ({
                  after: endCursor,
                  append: true,
                  id: current.id + 1,
                }));
              }}
            >
              {t("detailsPage.linear.loadMore")}
            </Button>
          </div>
        )}
      </ComboboxPopup>
    </Combobox>
  );
}
