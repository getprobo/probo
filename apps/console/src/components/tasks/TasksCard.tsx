// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatDate, formatDuration, formatError, promisifyMutation } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Card,
  DropdownItem,
  IconArrowCornerDownLeft,
  IconPencil,
  IconTrashCan,
  PriorityLevel,
  Spinner,
  TabBadge,
  TabItem,
  Tabs,
  TaskStateIcon,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { Fragment, type ReactNode, useState, useTransition } from "react";
import {
  graphql,
  readInlineData,
  useFragment,
  useMutation,
  useRefetchableFragment,
  useRelayEnvironment,
} from "react-relay";
import { Link, useLocation, useParams } from "react-router";

import type { TaskFormDialogFragment$key } from "#/__generated__/core/TaskFormDialogFragment.graphql";
import type {
  TaskFormDialogUpdateMutation,
  TaskPriority,
} from "#/__generated__/core/TaskFormDialogUpdateMutation.graphql";
import type { TasksCard_task$key } from "#/__generated__/core/TasksCard_task.graphql";
import type { TasksCard_TaskRowFragment$key } from "#/__generated__/core/TasksCard_TaskRowFragment.graphql";
import type { TasksCardDeleteMutation } from "#/__generated__/core/TasksCardDeleteMutation.graphql";
import type {
  TasksCardOrganizationFragment$data,
  TasksCardOrganizationFragment$key,
} from "#/__generated__/core/TasksCardOrganizationFragment.graphql";
import type { TasksCardOrganizationQuery } from "#/__generated__/core/TasksCardOrganizationQuery.graphql";
import TaskFormDialog, {
  taskPriorities,
  taskUpdateMutation,
} from "#/components/tasks/TaskFormDialog";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";

function resolveDropPriority(
  dragged: TaskPriority,
  above?: TaskPriority,
  below?: TaskPriority,
): TaskPriority | undefined {
  // If any neighbor shares the dragged priority, keep it.
  if (above === dragged || below === dragged) return undefined;

  // At edges, take the single neighbor's priority.
  if (!above && below) return below !== dragged ? below : undefined;
  if (!below && above) return above !== dragged ? above : undefined;

  // Both neighbors differ — pick the one closest to dragged.
  if (above && below) {
    const di = taskPriorities.indexOf(dragged);
    const dAbove = Math.abs(taskPriorities.indexOf(above) - di);
    const dBelow = Math.abs(taskPriorities.indexOf(below) - di);
    return dAbove <= dBelow ? above : below;
  }

  return undefined;
}

type Props = {
  tasks: TasksCardOrganizationFragment$data["tasks"]["edges"];
  connectionId: string;
  canReorder?: boolean;
  refetch?: (vars: Record<string, never>, options?: { fetchPolicy?: "store-and-network" | "network-only" }) => void;
};

const taskInlineFragment = graphql`
  fragment TasksCard_task on Task @inline {
    id
    state
    priority
    rank
  }
`;

function readTask(key: TasksCard_task$key) {
  return readInlineData(taskInlineFragment, key);
}

const organizationTasksFragment = graphql`
  fragment TasksCardOrganizationFragment on Organization
  @refetchable(queryName: "TasksCardOrganizationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    order: { type: "TaskOrder", defaultValue: { field: PRIORITY_RANK, direction: ASC } }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    canCreateTask: permission(action: "core:task:create")
    canUpdateTask: permission(action: "core:task:update")
    tasks(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "TasksCardOrganization_tasks") @required(action: THROW) {
      __id
      edges @required(action: THROW) {
        node {
          ...TasksCard_task
          ...TaskFormDialogFragment
          ...TasksCard_TaskRowFragment
        }
      }
    }
  }
`;

type OrganizationTasksCardProps = {
  organizationRef: TasksCardOrganizationFragment$key;
  header?: (params: { connectionId: string; canCreateTask: boolean; refetch: () => void }) => ReactNode;
};

export function OrganizationTasksCard({ organizationRef, header }: OrganizationTasksCardProps) {
  const [data, refetch] = useRefetchableFragment<
    TasksCardOrganizationQuery,
    TasksCardOrganizationFragment$key
  >(organizationTasksFragment, organizationRef);

  const handleRefetch = () => {
    refetch({}, { fetchPolicy: "store-and-network" });
  };

  return (
    <>
      {header?.({ connectionId: data.tasks.__id, canCreateTask: data.canCreateTask, refetch: handleRefetch })}
      <TasksCard
        tasks={data.tasks.edges}
        connectionId={data.tasks.__id}
        canReorder={data.canUpdateTask}
        refetch={refetch}
      />
    </>
  );
}

const updateRankMutation = graphql`
  mutation TasksCardUpdateRankMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        id
        priority
        rank
      }
    }
  }
`;

export function TasksCard({ tasks, connectionId, canReorder, refetch }: Props) {
  const { __ } = useTranslate();
  const hash = useLocation().hash.replace("#", "");
  const [, startTransition] = useTransition();

  const { toast } = useToast();
  const [draggedId, setDraggedId] = useState<string | null>(null);
  const [previewOrder, setPreviewOrder] = useState<string[] | null>(null);
  const [updateRank] = useMutation<TaskFormDialogUpdateMutation>(updateRankMutation);

  const handleStateChange = () => {
    if (refetch) {
      startTransition(() => {
        refetch({}, { fetchPolicy: "store-and-network" });
      });
    }
  };

  const hashes = [
    { hash: "", label: __("To do"), state: "TODO" },
    { hash: "done", label: __("Done"), state: "DONE" },
    { hash: "all", label: __("All"), state: null },
  ] as const;

  const tasksPerHash = new Map([
    ["", tasks?.filter(({ node }) => readTask(node).state === "TODO")],
    ["done", tasks?.filter(({ node }) => readTask(node).state === "DONE")],
    ["all", tasks],
  ]);

  const filteredTasks = tasksPerHash.get(hash) ?? [];
  const canDrag = !!canReorder && hash !== "all";

  const handleDragOver = (e: React.DragEvent, hoveredId: string) => {
    e.preventDefault();
    if (draggedId === null || hoveredId === draggedId) return;
    const ids = filteredTasks.map(({ node }) => readTask(node).id);
    const fromIdx = ids.indexOf(draggedId);
    if (fromIdx === -1) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const midY = rect.top + rect.height / 2;
    const insertBefore = e.clientY < midY;
    const hoverIdx = ids.indexOf(hoveredId);
    let targetIdx = insertBefore ? hoverIdx : hoverIdx + 1;
    if (targetIdx > fromIdx) targetIdx--;
    if (targetIdx === fromIdx) {
      setPreviewOrder(null);
      return;
    }
    const reordered = [...ids];
    reordered.splice(fromIdx, 1);
    reordered.splice(targetIdx, 0, draggedId);
    setPreviewOrder(reordered);
  };

  const handleDrop = () => {
    if (draggedId === null || previewOrder === null) {
      setDraggedId(null);
      return;
    }

    const newIdx = previewOrder.indexOf(draggedId);
    const originalIds = filteredTasks.map(({ node }) => readTask(node).id);
    const originalIdx = originalIds.indexOf(draggedId);
    if (originalIdx === -1) {
      setDraggedId(null);
      setPreviewOrder(null);
      return;
    }
    let targetOriginalIdx = newIdx;
    if (targetOriginalIdx >= originalIdx) targetOriginalIdx++;
    if (targetOriginalIdx >= filteredTasks.length) targetOriginalIdx = filteredTasks.length - 1;
    const targetTask = readTask(filteredTasks[targetOriginalIdx].node);
    const draggedTask = readTask(filteredTasks[originalIdx].node);

    // Determine target priority from neighbors at the drop position.
    const aboveId = newIdx > 0 ? previewOrder[newIdx - 1] : null;
    const belowId = newIdx < previewOrder.length - 1 ? previewOrder[newIdx + 1] : null;
    const aboveTask = aboveId ? readTask(filteredTasks[originalIds.indexOf(aboveId)].node) : null;
    const belowTask = belowId ? readTask(filteredTasks[originalIds.indexOf(belowId)].node) : null;
    const targetPriority = resolveDropPriority(draggedTask.priority, aboveTask?.priority, belowTask?.priority);

    setDraggedId(null);

    updateRank({
      variables: {
        input: {
          taskId: draggedId,
          rank: targetTask.rank,
          ...(targetPriority && { priority: targetPriority }),
        },
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to reorder task."),
              errors,
            ),
            variant: "error",
          });
        }
        if (refetch) {
          startTransition(() => {
            refetch(
              {},
              { fetchPolicy: errors?.length ? "network-only" : "store-and-network" },
            );
          });
        }
      },
      onError: () => {
        toast({
          title: __("Error"),
          description: __("Failed to reorder task."),
          variant: "error",
        });
      },
    });
  };

  const displayTasks = (() => {
    if (!previewOrder) return filteredTasks;
    const byId = new Map(filteredTasks.map(edge => [readTask(edge.node).id, edge]));
    const currentIdSet = new Set(byId.keys());
    const previewIdSet = new Set(previewOrder);
    if (currentIdSet.size !== previewIdSet.size || [...currentIdSet].some(id => !previewIdSet.has(id))) {
      return filteredTasks;
    }
    return previewOrder.map(id => byId.get(id)!);
  })();

  return (
    <div className="space-y-6">
      {tasks.length === 0
        ? (
            <p className="text-center py-6 text-txt-secondary">{__("No tasks")}</p>
          )
        : (
            <Card>
              <Tabs className="px-6">
                {hashes.map(h => (
                  <TabItem asChild active={hash === h.hash} key={h.hash}>
                    <Link to={`#${h.hash}`}>
                      {h.label}
                      <TabBadge>{tasksPerHash.get(h.hash)?.length}</TabBadge>
                    </Link>
                  </TabItem>
                ))}
              </Tabs>
              <div className="divide-y divide-border-solid">
                {hash === "all"
                  // All tabs group the todo using the state
                  ? hashes
                      .slice(0, 2)
                      .filter(h => tasksPerHash.get(h.hash)?.length)
                      .map(h => (
                        <Fragment key={h.label}>
                          <h2 className="px-6 py-3 text-sm font-medium flex items-center gap-2 bg-subtle">
                            <TaskStateIcon state={h.state!} />
                            {h.label}
                          </h2>
                          {tasksPerHash.get(h.hash)?.map(({ node }) => (
                            <TaskRow
                              key={readTask(node).id}
                              fKey={node}
                              connectionId={connectionId}
                              onStateChange={handleStateChange}
                            />
                          ))}
                        </Fragment>
                      ))
                  // Todo and Done tab simply list todos
                  : displayTasks.map(({ node }) => {
                      const task = readTask(node);
                      return (
                        <TaskRow
                          key={task.id}
                          fKey={node}
                          connectionId={connectionId}
                          canDrag={canDrag}
                          isDragging={draggedId === task.id}
                          isGhost={previewOrder !== null && draggedId === task.id}
                          onDragStart={() => setDraggedId(task.id)}
                          onDragOver={e => handleDragOver(e, task.id)}
                          onDrop={handleDrop}
                          onDragEnd={() => setDraggedId(null)}
                          onStateChange={handleStateChange}
                        />
                      );
                    })}
              </div>
            </Card>
          )}
      {canDrag && filteredTasks.length > 1 && (
        <p className="text-sm text-txt-tertiary">
          {__("Drag and drop to reorder tasks")}
        </p>
      )}
    </div>
  );
}

type TaskRowProps = {
  fKey: TasksCard_TaskRowFragment$key | TaskFormDialogFragment$key;
  connectionId: string;
  canDrag?: boolean;
  isDragging?: boolean;
  isGhost?: boolean;
  onDragStart?: () => void;
  onDragOver?: (e: React.DragEvent) => void;
  onDrop?: () => void;
  onDragEnd?: () => void;
  onStateChange?: () => void;
};

const fragment = graphql`
  fragment TasksCard_TaskRowFragment on Task {
    id
    name
    state
    priority
    description
    timeEstimate
    deadline
    canUpdate: permission(action: "core:task:update")
    canDelete: permission(action: "core:task:delete")
    assignedTo {
      id
      fullName
    }
    measure {
      id
      name
    }
  }
`;

const deleteMutation = graphql`
  mutation TasksCardDeleteMutation(
    $input: DeleteTaskInput!
    $connections: [ID!]!
  ) {
    deleteTask(input: $input) {
      deletedTaskId @deleteEdge(connections: $connections)
    }
  }
`;

function TaskRow(props: TaskRowProps) {
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [deleteTask] = useMutation<TasksCardDeleteMutation>(deleteMutation);
  const params = useParams<{ measureId?: string }>();

  const relayEnv = useRelayEnvironment();
  const { canUpdate, canDelete, ...task }
    = useFragment<TasksCard_TaskRowFragment$key>(
      fragment,
      props.fKey as TasksCard_TaskRowFragment$key,
    );
  const [updateTask, isUpdating] = useMutation<TaskFormDialogUpdateMutation>(taskUpdateMutation);

  const [isMouseDown, setIsMouseDown] = useState(false);

  const onToggle = async () => {
    await promisifyMutation(updateTask)({
      variables: {
        input: {
          taskId: task.id,
          state: task.state === "TODO" ? "DONE" : "TODO",
        },
      },
    });
    props.onStateChange?.();
  };

  const onDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteTask)({
          variables: {
            input: { taskId: task.id },
            connections: [props.connectionId],
          },
          onCompleted: (_response, errors) => {
            if (!errors && params.measureId) {
              updateStoreCounter(
                relayEnv,
                params.measureId,
                "tasks(first:0)",
                -1,
              );
            }
          },
        }),
      {
        message: "Are you sure you want to delete this task?",
      },
    );
  };

  const canDrag = props.canDrag;
  const isDragging = props.isDragging;
  const isGhost = props.isGhost;

  const className = [
    "transition-all duration-150",
    canDrag && isDragging && !isGhost && "opacity-40 cursor-grabbing",
    canDrag && !isDragging && !isMouseDown && "cursor-grab",
    canDrag && !isDragging && isMouseDown && "cursor-grabbing",
    isGhost && "opacity-50 bg-primary-50",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <TaskFormDialog
        task={props.fKey as TaskFormDialogFragment$key}
        ref={dialogRef}
        onCompleted={props.onStateChange}
      />
      <div
        className={`flex items-center justify-between py-3 px-6 ${className}`}
        draggable={canDrag}
        onDragStart={canDrag ? props.onDragStart : undefined}
        onDragOver={canDrag ? props.onDragOver : undefined}
        onDrop={canDrag ? props.onDrop : undefined}
        onDragEnd={canDrag ? props.onDragEnd : undefined}
        onMouseDown={canDrag ? () => setIsMouseDown(true) : undefined}
        onMouseUp={canDrag ? () => setIsMouseDown(false) : undefined}
        onMouseLeave={canDrag ? () => setIsMouseDown(false) : undefined}
      >
        <div className="flex gap-2 items-start">
          <div className="flex items-center gap-2 pt-[2px]">
            <PriorityLevel level={task.priority} />
            <button
              onClick={() => void onToggle()}
              className="cursor-pointer -m-1 p-1 disabled:opacity-60"
              disabled={isUpdating}
            >
              <TaskStateIcon state={task.state} />
            </button>
          </div>
          <div className="text-sm space-y-1 flex-1">
            <h2 className="font-medium">{task.name}</h2>
            {task.description && (
              <p className="text-txt-secondary whitespace-pre-wrap wrap-break-word">
                {task.description}
              </p>
            )}

            <div className="flex flex-wrap items-center gap-3 text-txt-secondary text-xs">
              {task.measure && (
                <span className="flex items-center gap-1">
                  <IconArrowCornerDownLeft className="scale-x-[-1]" size={14} />
                  <Link
                    className="hover:underline"
                    to={`/organizations/${organizationId}/measures/${task.measure?.id}`}
                  >
                    {task.measure?.name}
                  </Link>
                </span>
              )}
              {task.timeEstimate && (
                <span>{formatDuration(task.timeEstimate, __)}</span>
              )}
              {task.deadline && (
                <time dateTime={task.deadline}>
                  {formatDate(task.deadline)}
                </time>
              )}
            </div>
          </div>
        </div>
        {task.assignedTo?.fullName && (
          <div className="text-sm text-txt-secondary ml-auto mr-8">
            <Link
              className="hover:underline"
              to={`/organizations/${organizationId}/people/${task.assignedTo.id}`}
            >
              {task.assignedTo.fullName}
            </Link>
          </div>
        )}
        <div className="flex gap-2 items-center">
          {isUpdating && <Spinner size={16} />}
          {(canUpdate || canDelete) && (
            <ActionDropdown>
              {canUpdate && (
                <DropdownItem
                  icon={IconPencil}
                  onClick={() => dialogRef.current?.open()}
                >
                  {__("Edit")}
                </DropdownItem>
              )}
              {canDelete && (
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={onDelete}
                >
                  {__("Delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          )}
        </div>
      </div>
    </>
  );
}
