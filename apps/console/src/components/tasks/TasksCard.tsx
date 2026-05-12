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
  Button,
  Card,
  IconArrowCornerDownLeft,
  IconCircleCheck,
  IconCircleProgress,
  IconPencil,
  IconTrashCan,
  PriorityLevel,
  TabBadge,
  TabItem,
  Tabs,
  TaskStateIcon,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { Fragment, type ReactNode, useRef, useState, useTransition } from "react";
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
  if (above === dragged || below === dragged) return undefined;

  const di = taskPriorities.indexOf(dragged);

  if (!above && below) {
    return taskPriorities.indexOf(below) <= di ? below : undefined;
  }
  if (!below && above) {
    return taskPriorities.indexOf(above) >= di ? above : undefined;
  }
  if (above && below) {
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
        state
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
  const [dropTargetState, setDropTargetState] = useState<string | null>(null);
  const [updateRank] = useMutation<TaskFormDialogUpdateMutation>(updateRankMutation);
  const droppedRef = useRef(false);

  const handleStateChange = () => {
    if (refetch) {
      startTransition(() => {
        refetch({}, { fetchPolicy: "store-and-network" });
      });
    }
  };

  const stateHashes = [
    { hash: "todo", label: __("To do"), state: "TODO" },
    { hash: "in-progress", label: __("In progress"), state: "IN_PROGRESS" },
    { hash: "done", label: __("Done"), state: "DONE" },
  ] as const;

  const hashes = [
    { hash: "", label: __("All"), state: null },
    ...stateHashes,
  ] as const;

  const tasksPerHash = new Map<string, typeof tasks>([
    ...stateHashes.map(h => [h.hash, tasks?.filter(({ node }) => readTask(node).state === h.state)] as const),
    ["", tasks],
  ]);

  const filteredTasks = tasksPerHash.get(hash) ?? [];
  const canDrag = !!canReorder;

  // Get the task list for a given state section.
  const sectionTasks = (state: string) =>
    tasks?.filter(({ node }) => readTask(node).state === state) ?? [];

  const handleDragOver = (e: React.DragEvent, hoveredId: string, hoveredState?: string) => {
    e.preventDefault();
    if (draggedId === null || hoveredId === draggedId) return;

    if (hoveredState) setDropTargetState(hoveredState);

    // Reorder within the target section (works for both All and single-state tabs).
    const sectionList = hash === "" && hoveredState
      ? sectionTasks(hoveredState)
      : filteredTasks;

    const ids = sectionList.map(({ node }) => readTask(node).id);
    const fromIdx = ids.indexOf(draggedId);
    const rect = e.currentTarget.getBoundingClientRect();
    const midY = rect.top + rect.height / 2;
    const insertBefore = e.clientY < midY;
    const hoverIdx = ids.indexOf(hoveredId);

    if (fromIdx === -1) {
      // Dragging from another section — insert relative to the hovered task.
      const targetIdx = insertBefore ? hoverIdx : hoverIdx + 1;
      const reordered = [...ids];
      reordered.splice(targetIdx, 0, draggedId);
      setPreviewOrder(reordered);
      return;
    }

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
    if (draggedId === null) {
      resetDragState();
      return;
    }

    if (previewOrder === null && !(hash === "" && dropTargetState)) {
      resetDragState();
      return;
    }

    // Determine which section list to resolve rank/priority from.
    const targetState = hash === "" ? dropTargetState : null;
    const sectionList = targetState ? sectionTasks(targetState) : filteredTasks;
    const sectionIds = sectionList.map(({ node }) => readTask(node).id);
    const byId = new Map(tasks.map(edge => [readTask(edge.node).id, edge]));

    // Use previewOrder when available, otherwise the section list.
    // Append draggedId if missing (cross-section drop onto a header with no preview).
    const order = previewOrder ?? (sectionIds.includes(draggedId) ? sectionIds : [...sectionIds, draggedId]);
    const newIdx = order.indexOf(draggedId);

    if (newIdx === -1) {
      resetDragState();
      return;
    }

    // Find the task we're displacing to get its rank.
    const originalIdx = sectionIds.indexOf(draggedId);
    let targetOriginalIdx = newIdx;
    if (originalIdx !== -1) {
      if (targetOriginalIdx >= originalIdx) targetOriginalIdx++;
      if (targetOriginalIdx >= sectionList.length) targetOriginalIdx = sectionList.length - 1;
    } else {
      // Cross-section drop: clamp to section bounds.
      if (targetOriginalIdx >= sectionList.length) targetOriginalIdx = sectionList.length - 1;
    }

    const draggedEdge = byId.get(draggedId);
    if (!draggedEdge) {
      resetDragState();
      return;
    }
    const draggedTask = readTask(draggedEdge.node);

    // Determine target rank from the displaced task, or default to rank 1 for empty sections.
    const targetRank = sectionList.length > 0
      ? readTask(sectionList[Math.max(0, targetOriginalIdx)].node).rank
      : 1;

    // Determine if state changed (All tab cross-section drop).
    const newState = targetState && targetState !== draggedTask.state
      ? targetState as "TODO" | "IN_PROGRESS" | "DONE"
      : undefined;

    // Only change priority for same-state reorder, never for cross-section drops.
    const aboveId = newIdx > 0 ? order[newIdx - 1] : null;
    const belowId = newIdx < order.length - 1 ? order[newIdx + 1] : null;
    const aboveTask = aboveId && byId.has(aboveId) ? readTask(byId.get(aboveId)!.node) : null;
    const belowTask = belowId && byId.has(belowId) ? readTask(byId.get(belowId)!.node) : null;
    const targetPriority = newState
      ? undefined
      : resolveDropPriority(draggedTask.priority, aboveTask?.priority, belowTask?.priority);

    const taskId = draggedId;

    droppedRef.current = true;

    updateRank({
      variables: {
        input: {
          taskId,
          rank: targetRank,
          ...(targetPriority && { priority: targetPriority }),
          ...(newState && { state: newState }),
        },
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to reorder task."), errors),
            variant: "error",
          });
        }
        if (refetch) {
          startTransition(() => {
            refetch({}, { fetchPolicy: errors?.length ? "network-only" : "store-and-network" });
            droppedRef.current = false;
            resetDragState();
          });
        } else {
          droppedRef.current = false;
          resetDragState();
        }
      },
      onError: () => {
        droppedRef.current = false;
        resetDragState();
        toast({ title: __("Error"), description: __("Failed to reorder task."), variant: "error" });
      },
    });
  };

  const resetDragState = () => {
    setDraggedId(null);
    setPreviewOrder(null);
    setDropTargetState(null);
  };

  const byId = new Map(tasks.map(edge => [readTask(edge.node).id, edge]));

  const applyPreviewOrder = (sourceTasks: typeof tasks) => {
    if (!previewOrder) return sourceTasks;
    const sourceIds = new Set(sourceTasks.map(({ node }) => readTask(node).id));
    // Check if the preview order matches this section (may include the dragged item from another section).
    const previewMatchesSection = previewOrder.every(id => sourceIds.has(id) || id === draggedId);
    if (!previewMatchesSection) return sourceTasks;
    return previewOrder.filter(id => byId.has(id)).map(id => byId.get(id)!);
  };

  const displayTasks = applyPreviewOrder(filteredTasks);

  const renderTaskRow = (node: (typeof tasks)[number]["node"], sectionState?: "TODO" | "IN_PROGRESS" | "DONE") => {
    const task = readTask(node);
    return (
      <TaskRow
        key={task.id}
        fKey={node}
        connectionId={connectionId}
        sectionState={sectionState}
        canDrag={canDrag}
        isDragging={draggedId === task.id}
        isGhost={previewOrder !== null && draggedId === task.id}
        onDragStart={() => setDraggedId(task.id)}
        onDragOver={e => handleDragOver(e, task.id, sectionState)}
        onDrop={handleDrop}
        onDragEnd={() => { if (!droppedRef.current) resetDragState(); }}
        onStateChange={handleStateChange}
      />
    );
  };

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
                      {h.state && <TaskStateIcon state={h.state} />}
                      {h.label}
                      <TabBadge>{tasksPerHash.get(h.hash)?.length}</TabBadge>
                    </Link>
                  </TabItem>
                ))}
              </Tabs>
              <div className="divide-y divide-border-solid">
                {hash === ""
                  ? stateHashes
                      .filter(h => tasksPerHash.get(h.hash)?.length || (draggedId && dropTargetState === h.state))
                      .map((h) => {
                        const displayEdges = applyPreviewOrder(tasksPerHash.get(h.hash) ?? []);
                        const dragClass = canDrag && draggedId !== null
                          ? "border-2 border-dashed border-transparent hover:border-primary-300"
                          : "";
                        return (
                          <Fragment key={h.label}>
                            <h2
                              className={`px-6 py-3 text-sm font-medium flex items-center gap-2 bg-subtle ${dragClass}`}
                              onDragOver={canDrag
                                ? (e) => {
                                    e.preventDefault();
                                    setDropTargetState(h.state);
                                  }
                                : undefined}
                              onDrop={canDrag ? handleDrop : undefined}
                            >
                              <TaskStateIcon state={h.state} />
                              {h.label}
                            </h2>
                            {displayEdges.map(({ node }) => renderTaskRow(node, h.state))}
                          </Fragment>
                        );
                      })
                  : displayTasks.map(({ node }) => renderTaskRow(node))}
              </div>
            </Card>
          )}
      {canDrag && filteredTasks.length > 1 && (
        <p className="text-sm text-txt-tertiary">
          {hash === ""
            ? __("Drag and drop to reorder tasks or move them between states")
            : __("Drag and drop to reorder tasks")}
        </p>
      )}
    </div>
  );
}

type TaskRowProps = {
  fKey: TasksCard_TaskRowFragment$key | TaskFormDialogFragment$key;
  connectionId: string;
  sectionState?: "TODO" | "IN_PROGRESS" | "DONE";
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
  const [updateTask, isAdvancing] = useMutation<TaskFormDialogUpdateMutation>(taskUpdateMutation);
  const [isMouseDown, setIsMouseDown] = useState(false);
  const displayState = props.sectionState ?? task.state;

  const nextStepConfig: Record<string, {
    state: "IN_PROGRESS" | "DONE";
    label: string;
    icon: typeof IconCircleProgress;
    className: string;
  }> = {
    TODO: { state: "IN_PROGRESS", label: __("Move to In progress"), icon: IconCircleProgress, className: "text-txt-warning" },
    IN_PROGRESS: { state: "DONE", label: __("Move to Done"), icon: IconCircleCheck, className: "text-txt-accent" },
  };

  const onAdvance = async () => {
    const config = nextStepConfig[displayState];
    if (!config) return;
    const target = config.state;
    await promisifyMutation(updateTask)({
      variables: {
        input: {
          taskId: task.id,
          state: target,
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

  const { canDrag, isDragging, isGhost } = props;

  const className = [
    canDrag && "select-none",
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
            <TaskStateIcon state={displayState} />
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
          {canUpdate && nextStepConfig[displayState] && (
            <Button
              variant="secondary"
              icon={nextStepConfig[displayState].icon}
              className={nextStepConfig[displayState].className}
              title={nextStepConfig[displayState].label}
              onClick={() => void onAdvance()}
              disabled={isAdvancing}
            />
          )}
          {canUpdate && (
            <Button
              variant="secondary"
              icon={IconPencil}
              title={__("Edit")}
              onClick={() => dialogRef.current?.open()}
            />
          )}
          {canDelete && (
            <Button
              variant="danger"
              icon={IconTrashCan}
              title={__("Delete")}
              onClick={onDelete}
            />
          )}
        </div>
      </div>
    </>
  );
}
