// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  IconChevronDown,
  IconPlusLarge,
  IconTrashCan,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery, useRefetchableFragment } from "react-relay";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { InternalControlTasksTab_internalControl$key } from "#/__generated__/core/InternalControlTasksTab_internalControl.graphql";
import type { InternalControlTasksTabQuery } from "#/__generated__/core/InternalControlTasksTabQuery.graphql";
import type { InternalControlTasksTabRefetchQuery } from "#/__generated__/core/InternalControlTasksTabRefetchQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CreateTaskDialog } from "#/pages/organizations/tasks/_components/CreateTaskDialog";
import { TaskStateIcon } from "#/pages/organizations/tasks/_components/TaskStateIcon";
import { emptyTaskListFilter, taskDetailsPath } from "#/pages/organizations/tasks/_lib/taskPath";
import { taskStateKeys } from "#/pages/organizations/tasks/_lib/taskState";
import { useTaskInternalControlIds } from "#/pages/organizations/tasks/_lib/useTaskInternalControlIds";
import { useUpdateTask } from "#/pages/organizations/tasks/_lib/useUpdateTask";

import { InternalControlLinkTasksDialog } from "./InternalControlLinkTasksDialog";
import { internalControlTasksTab } from "./variants";

const previewLimit = 4;

const internalControlTasksTabFragment = graphql`
  fragment InternalControlTasksTab_internalControl on InternalControl
  @refetchable(queryName: "InternalControlTasksTabRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    after: { type: "CursorKey", defaultValue: null }
    filter: { type: "TaskFilter", defaultValue: null }
  ) {
    canCreateTask: permission(action: "core:task:create")
    canUpdateTask: permission(action: "core:task:update")
    tasks(
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: $filter
    ) @connection(key: "InternalControl__tasks") @required(action: THROW) {
      __id
      edges @required(action: THROW) {
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

const tasksQuery = graphql`
  query InternalControlTasksTabQuery($internalControlId: ID!, $filter: TaskFilter) {
    node(id: $internalControlId) @required(action: THROW) {
      __typename
      ... on InternalControl {
        ...InternalControlTasksTab_internalControl @arguments(filter: $filter)
      }
    }
  }
`;

interface InternalControlTasksPanelProps {
  internalControlKey: InternalControlTasksTab_internalControl$key;
  internalControlId: string;
}

function InternalControlTasksPanel({
  internalControlKey,
  internalControlId,
}: InternalControlTasksPanelProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const [internalControl] = useRefetchableFragment<
    InternalControlTasksTabRefetchQuery,
    InternalControlTasksTab_internalControl$key
  >(internalControlTasksTabFragment, internalControlKey);
  const tasks = internalControl.tasks.edges.map(edge => edge.node);
  const [limit, setLimit] = useState<number | null>(previewLimit);
  const visibleTasks = limit == null ? tasks : tasks.slice(0, limit);
  const hiddenCount = limit == null ? 0 : Math.max(tasks.length - limit, 0);
  const [updateTask, isUpdating] = useUpdateTask();
  const { root, state, actions, actionCell, action } = internalControlTasksTab();
  const columnCount = internalControl.canUpdateTask ? 3 : 2;
  const showActions = internalControl.canCreateTask || internalControl.canUpdateTask;

  return (
    <div className={root()}>
      <Table>
        <Thead>
          <Tr>
            <Th>{t("internalControlTasksTab.columns.name")}</Th>
            <Th>{t("internalControlTasksTab.columns.state")}</Th>
            {internalControl.canUpdateTask && <Th />}
          </Tr>
        </Thead>
        <Tbody>
          {tasks.length === 0 && (
            <Tr>
              <Td colSpan={columnCount} className="text-center text-txt-secondary">
                {t("internalControlTasksTab.empty")}
              </Td>
            </Tr>
          )}
          {visibleTasks.map(task => (
            <Tr key={task.id} to={taskDetailsPath(organizationId, task.id)}>
              <Td>{task.name}</Td>
              <Td>
                <span className={state()}>
                  <TaskStateIcon state={task.state} />
                  {t(`tasksCard.states.${taskStateKeys[task.state]}`)}
                </span>
              </Td>
              {internalControl.canUpdateTask && (
                <Td noLink width={50} className="text-end">
                  {task.canUpdate && (
                    <UnlinkTaskButton
                      taskKey={task}
                      internalControlId={internalControlId}
                      disabled={isUpdating}
                      onUnlink={updateTask}
                    />
                  )}
                </Td>
              )}
            </Tr>
          ))}
          {showActions && (
            <tr>
              <td colSpan={columnCount}>
                <div className={actions()}>
                  {internalControl.canCreateTask && (
                    <div className={actionCell()}>
                      <CreateTaskDialog
                        connectionId={internalControl.tasks.__id}
                        internalControlId={internalControlId}
                      >
                        <button type="button" className={action()} disabled={isUpdating}>
                          <IconPlusLarge size={16} />
                          {t("internalControlTasksTab.actions.newTask")}
                        </button>
                      </CreateTaskDialog>
                    </div>
                  )}
                  {internalControl.canUpdateTask && (
                    <div className={actionCell()}>
                      <InternalControlLinkTasksDialog
                        organizationId={organizationId}
                        internalControlId={internalControlId}
                        disabled={isUpdating}
                      >
                        <button type="button" className={action()} disabled={isUpdating}>
                          <IconPlusLarge size={16} />
                          {t("internalControlTasksTab.actions.link")}
                        </button>
                      </InternalControlLinkTasksDialog>
                    </div>
                  )}
                </div>
              </td>
            </tr>
          )}
        </Tbody>
      </Table>
      {hiddenCount > 0 && (
        <Button
          variant="tertiary"
          onClick={() => setLimit(null)}
          className="mx-auto mt-3"
          icon={IconChevronDown}
        >
          {t("internalControlTasksTab.actions.showMore", { count: hiddenCount })}
        </Button>
      )}
    </div>
  );
}

function UnlinkTaskButton({
  taskKey,
  internalControlId,
  disabled,
  onUnlink,
}: {
  taskKey: Parameters<typeof useTaskInternalControlIds>[0];
  internalControlId: string;
  disabled?: boolean;
  onUnlink: ReturnType<typeof useUpdateTask>[0];
}) {
  const { t } = useTranslation();
  const { taskId, ids, pending, failed, retry } = useTaskInternalControlIds(taskKey);

  function unlink() {
    if (pending || failed) {
      return;
    }

    void onUnlink({
      previousInternalControlIds: ids,
      variables: {
        input: {
          taskId,
          internalControlIds: ids.filter(id => id !== internalControlId),
        },
      },
    });
  }

  if (failed) {
    return (
      <Button variant="secondary" disabled={disabled} onClick={retry}>
        {t("internalControlTasksTab.actions.retry")}
      </Button>
    );
  }

  return (
    <Button
      variant="secondary"
      disabled={disabled || pending}
      onClick={unlink}
      icon={IconTrashCan}
    >
      {t("internalControlTasksTab.actions.unlink")}
    </Button>
  );
}

export default function InternalControlTasksTab() {
  const { internalControlId } = useParams<{ internalControlId: string }>();
  if (!internalControlId) {
    throw new Error("Missing :internalControlId param in route");
  }

  return <InternalControlTasksQuery key={internalControlId} internalControlId={internalControlId} />;
}

interface InternalControlTasksQueryProps {
  internalControlId: string;
}

function InternalControlTasksQuery({ internalControlId }: InternalControlTasksQueryProps) {
  const { node } = useLazyLoadQuery<InternalControlTasksTabQuery>(
    tasksQuery,
    { internalControlId, filter: emptyTaskListFilter },
  );
  if (node.__typename !== "InternalControl") {
    throw new Error("invalid node type");
  }

  return (
    <InternalControlTasksPanel
      internalControlKey={node}
      internalControlId={internalControlId}
    />
  );
}
