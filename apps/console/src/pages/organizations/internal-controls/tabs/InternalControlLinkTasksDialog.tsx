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

import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  InfiniteScrollTrigger,
  Input,
  Spinner,
} from "@probo/ui";
import { type ReactNode, type RefObject, Suspense, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { useDebounceCallback } from "usehooks-ts";

import type { InternalControlLinkTasksDialog_organization$key } from "#/__generated__/core/InternalControlLinkTasksDialog_organization.graphql";
import type { InternalControlLinkTasksDialogPaginationQuery } from "#/__generated__/core/InternalControlLinkTasksDialogPaginationQuery.graphql";
import type { InternalControlLinkTasksDialogQuery } from "#/__generated__/core/InternalControlLinkTasksDialogQuery.graphql";
import { TaskStateIcon } from "#/pages/organizations/tasks/_components/TaskStateIcon";
import { type TaskState, taskStateKeys } from "#/pages/organizations/tasks/_lib/taskState";
import { useTaskInternalControlIds } from "#/pages/organizations/tasks/_lib/useTaskInternalControlIds";
import { useUpdateTask } from "#/pages/organizations/tasks/_lib/useUpdateTask";

import { internalControlTasksTab } from "./variants";

const pageSize = 20;

const tasksQuery = graphql`
  query InternalControlLinkTasksDialogQuery($organizationId: ID!, $first: Int!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...InternalControlLinkTasksDialog_organization @arguments(first: $first)
      }
    }
  }
`;

const tasksFragment = graphql`
  fragment InternalControlLinkTasksDialog_organization on Organization
  @refetchable(queryName: "InternalControlLinkTasksDialogPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    filter: { type: "TaskFilter", defaultValue: null }
  ) {
    tasks(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
      filter: $filter
    ) @connection(key: "InternalControlLinkTasksDialog_tasks", filters: ["filter"]) {
      edges {
        node {
          id
          name
          state
          canUpdate: permission(action: "core:task:update")
          ...useTaskInternalControlIds_task
        }
      }
    }
  }
`;

interface InternalControlLinkTasksDialogProps {
  children: ReactNode;
  organizationId: string;
  internalControlId: string;
  disabled?: boolean;
}

type SearchRef = RefObject<{ search: (value: string) => void } | null>;

export function InternalControlLinkTasksDialog({
  children,
  ...props
}: InternalControlLinkTasksDialogProps) {
  const { t } = useTranslation();
  const searchRef: SearchRef = useRef(null);

  return (
    <Dialog trigger={children} title={t("internalControlTasksTab.dialog.title")}>
      <DialogContent>
        <div className="sticky top-0 z-1 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6 py-4">
          <Input
            icon={IconMagnifyingGlass}
            placeholder={t("internalControlTasksTab.dialog.searchPlaceholder")}
            onValueChange={value => searchRef.current?.search(value)}
          />
        </div>
        <Suspense fallback={<Spinner centered />}>
          <InternalControlLinkTasksDialogContent {...props} ref={searchRef} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={t("internalControlTasksTab.dialog.actions.close")} />
    </Dialog>
  );
}

function InternalControlLinkTasksDialogContent({
  ref: searchRef,
  organizationId,
  internalControlId,
  disabled,
}: Omit<InternalControlLinkTasksDialogProps, "children"> & { ref: SearchRef }) {
  const { t } = useTranslation();
  const query = useLazyLoadQuery<InternalControlLinkTasksDialogQuery>(
    tasksQuery,
    { organizationId, first: pageSize },
    { fetchPolicy: "network-only" },
  );
  const { data, loadNext, hasNext, isLoadingNext, refetch } = usePaginationFragment<
    InternalControlLinkTasksDialogPaginationQuery,
    InternalControlLinkTasksDialog_organization$key
  >(
    tasksFragment,
    query.organization as InternalControlLinkTasksDialog_organization$key,
  );
  const tasks = data.tasks?.edges?.flatMap(edge => edge?.node ?? []) ?? [];
  const [updateTask, isUpdating] = useUpdateTask();
  const { state } = internalControlTasksTab();
  const handleSearch = useDebounceCallback((value: string) => {
    refetch({
      first: pageSize,
      filter: { query: value.trim() || null },
    });
  }, 300);

  useEffect(() => {
    searchRef.current = { search: handleSearch };
    return () => {
      searchRef.current = null;
    };
  }, [handleSearch, searchRef]);

  return (
    <div className="divide-y divide-border-low">
      {tasks.length === 0 && (
        <p className="px-6 py-4 text-txt-secondary">
          {t("internalControlTasksTab.dialog.empty")}
        </p>
      )}
      {tasks.map(task => (
        <TaskRow
          key={task.id}
          taskKey={task}
          name={task.name}
          state={task.state}
          stateClassName={state()}
          internalControlId={internalControlId}
          disabled={disabled || isUpdating || !task.canUpdate}
          onToggle={updateTask}
        />
      ))}
      {hasNext && (
        <InfiniteScrollTrigger
          loading={isLoadingNext}
          onView={() => loadNext(pageSize)}
        />
      )}
    </div>
  );
}

function TaskRow({
  taskKey,
  name,
  state,
  stateClassName,
  internalControlId,
  disabled,
  onToggle,
}: {
  taskKey: Parameters<typeof useTaskInternalControlIds>[0];
  name: string;
  state: TaskState;
  stateClassName: string;
  internalControlId: string;
  disabled?: boolean;
  onToggle: ReturnType<typeof useUpdateTask>[0];
}) {
  const { t } = useTranslation();
  const { taskId, ids, pending, failed, retry } = useTaskInternalControlIds(taskKey);
  const linked = ids.includes(internalControlId);
  const IconComponent = linked ? IconTrashCan : IconPlusLarge;
  const rowDisabled = disabled || pending;

  function toggleLink() {
    if (pending || failed) {
      return;
    }

    const nextInternalControlIds = linked
      ? ids.filter(id => id !== internalControlId)
      : [...ids, internalControlId];
    void onToggle({
      previousInternalControlIds: ids,
      variables: {
        input: {
          taskId,
          internalControlIds: nextInternalControlIds,
        },
      },
    });
  }

  return (
    <button
      type="button"
      className="flex w-full cursor-pointer items-center gap-4 px-6 py-4 hover:bg-subtle"
      onClick={failed ? retry : toggleLink}
      disabled={rowDisabled}
    >
      <span className="min-w-0 flex-1 truncate text-left">{name}</span>
      <span className={stateClassName}>
        <TaskStateIcon state={state} />
        {t(`tasksCard.states.${taskStateKeys[state]}`)}
      </span>
      <Button
        disabled={rowDisabled}
        className="ml-auto"
        variant={linked ? "secondary" : "primary"}
        asChild
      >
        <span>
          {failed
            ? t("internalControlTasksTab.actions.retry")
            : (
                <>
                  <IconComponent size={16} />
                  {" "}
                  {linked
                    ? t("internalControlTasksTab.dialog.actions.unlink")
                    : t("internalControlTasksTab.dialog.actions.link")}
                </>
              )}
        </span>
      </Button>
    </button>
  );
}
