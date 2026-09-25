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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { formatError } from "@probo/helpers";
import { useToast } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { TooltipProvider } from "@probo/ui/src/v2/Tooltip/TooltipProvider";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Fragment, type ReactNode, useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  readInlineData,
  useRefetchableFragment,
  useRelayEnvironment,
} from "react-relay";

import type { TasksCard_task$key } from "#/__generated__/core/TasksCard_task.graphql";
import type {
  TasksCardOrganizationFragment$data,
  TasksCardOrganizationFragment$key,
} from "#/__generated__/core/TasksCardOrganizationFragment.graphql";
import type { TasksCardOrganizationQuery } from "#/__generated__/core/TasksCardOrganizationQuery.graphql";
import type { TasksCardUpdateRankMutation } from "#/__generated__/core/TasksCardUpdateRankMutation.graphql";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { TaskStateIcon } from "#/pages/organizations/tasks/_components/TaskStateIcon";
import { insertNextTaskEdge } from "#/pages/organizations/tasks/_lib/taskConnectionOrder";
import {
  taskPriorities,
  type TaskPriority,
  type TaskState,
  taskStateKeys,
  taskStates,
} from "#/pages/organizations/tasks/_lib/taskState";
import {
  type TasksCardFilterInput,
  useTasksCardFilters,
} from "#/pages/organizations/tasks/_lib/useTasksCardFilters";
import { useTasksCardSearch } from "#/pages/organizations/tasks/_lib/useTasksCardSearch";

import { TaskListItem } from "./TaskListItem";
import { tasksCard } from "./variants";

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

interface TasksCardProps {
  tasks: TasksCardOrganizationFragment$data["tasks"]["edges"];
  connectionId: string;
  canReorder?: boolean;
  refetch?: (
    vars: { filter: TasksCardFilterInput },
    options?: {
      fetchPolicy?: "store-and-network" | "store-or-network" | "network-only";
    },
  ) => void;
}

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
    filter: { type: "TaskFilter", defaultValue: null }
  ) {
    canCreateTask: permission(action: "core:task:create")
    canUpdateTask: permission(action: "core:task:update")
    tasks(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "TasksCardOrganization_tasks") @required(action: THROW) {
      __id
      edges @required(action: THROW) {
        node {
          ...TasksCard_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

interface OrganizationTasksCardProps {
  organizationRef: TasksCardOrganizationFragment$key;
  header?: (params: { connectionId: string; canCreateTask: boolean; refetch: () => void }) => ReactNode;
}

export function OrganizationTasksCard({ organizationRef, header }: OrganizationTasksCardProps) {
  const { graphqlFilter } = useTasksCardFilters();
  const [data, refetch] = useRefetchableFragment<
    TasksCardOrganizationQuery,
    TasksCardOrganizationFragment$key
  >(organizationTasksFragment, organizationRef);

  const handleRefetch = () => {
    refetch({ filter: graphqlFilter }, { fetchPolicy: "store-and-network" });
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
        recurrenceInterval
        ...TasksCard_task
        ...TaskListItem_task
      }
      nextTaskEdge {
        node {
          ...TasksCard_task
          ...TaskListItem_task
          measure {
            id
          }
        }
      }
    }
  }
`;

export function TasksCard({ tasks, connectionId, canReorder, refetch }: TasksCardProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const relayEnv = useRelayEnvironment();
  const { query, state: selectedState, graphqlFilter, setState }
    = useTasksCardFilters();
  const [queryInput, setQueryInput] = useTasksCardSearch();
  const [, startTransition] = useTransition();
  const skipFirstFilterRefetch = useRef(true);

  const { toast } = useToast();
  const [draggedId, setDraggedId] = useState<string | null>(null);
  const [previewOrder, setPreviewOrder] = useState<string[] | null>(null);
  const [dropTargetState, setDropTargetState] = useState<string | null>(null);
  const [updateRank] = useMutation<TasksCardUpdateRankMutation>(
    updateRankMutation,
    { errorToast: false },
  );
  const droppedRef = useRef(false);
  const [ignoreClick, setIgnoreClick] = useState(false);

  const handleStateChange = () => {
    if (refetch) {
      startTransition(() => {
        refetch({ filter: graphqlFilter }, { fetchPolicy: "store-and-network" });
      });
    }
  };

  useEffect(() => {
    if (!refetch || skipFirstFilterRefetch.current) {
      skipFirstFilterRefetch.current = false;
      return;
    }

    startTransition(() => {
      refetch({ filter: graphqlFilter }, { fetchPolicy: "store-or-network" });
    });
  }, [graphqlFilter, refetch]);

  const tasksPerState = new Map<TaskState, typeof tasks>(
    taskStates.map(state => [
      state,
      tasks.filter(({ node }) => readTask(node).state === state),
    ]),
  );
  const filteredTasks = tasks;
  const canDrag = !!canReorder && query === "";
  const slots = tasksCard({ dragging: canDrag && draggedId !== null });

  // Get the task list for a given state section.
  const sectionTasks = (state: string) =>
    tasks.filter(({ node }) => readTask(node).state === state);

  const handleDragOver = (e: React.DragEvent, hoveredId: string, hoveredState?: string) => {
    e.preventDefault();
    if (draggedId === null || hoveredId === draggedId) return;

    if (hoveredState) setDropTargetState(hoveredState);

    // Reorder within the target section (works for both All and single-state tabs).
    const sectionList = selectedState == null && hoveredState
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

    if (previewOrder === null && !(selectedState == null && dropTargetState)) {
      resetDragState();
      return;
    }

    // Determine which section list to resolve rank/priority from.
    const targetState = selectedState == null ? dropTargetState : null;
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
      ? targetState as TaskState
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

    void updateRank({
      variables: {
        input: {
          taskId,
          rank: targetRank,
          ...(targetPriority && { priority: targetPriority }),
          ...(newState && { state: newState }),
        },
      },
      updater: (store) => {
        const spawnedMeasureId = insertNextTaskEdge(store, organizationId, [connectionId]);
        if (spawnedMeasureId) {
          updateStoreCounter(relayEnv, spawnedMeasureId, "tasks(first:0)", 1);
        }
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: t("tasksCard.error.title"),
            description: formatError(t("tasksCard.error.reorder"), errors),
            variant: "error",
          });
        }
        if (refetch) {
          startTransition(() => {
            refetch(
              { filter: graphqlFilter },
              { fetchPolicy: errors?.length ? "network-only" : "store-and-network" },
            );
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
        toast({
          title: t("tasksCard.error.title"),
          description: t("tasksCard.error.reorder"),
          variant: "error",
        });
      },
    }).catch(() => {
      // Error feedback is handled by the callbacks above.
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

  const renderTaskRow = (node: (typeof tasks)[number]["node"], sectionState?: TaskState) => {
    const task = readTask(node);
    return (
      <TaskListItem
        key={task.id}
        taskKey={node}
        connectionId={connectionId}
        sectionState={sectionState}
        canDrag={canDrag}
        isDragging={draggedId === task.id}
        isGhost={previewOrder !== null && draggedId === task.id}
        ignoreClick={ignoreClick}
        onDragStart={() => {
          setIgnoreClick(true);
          setDraggedId(task.id);
        }}
        onDragOver={e => handleDragOver(e, task.id, sectionState)}
        onDrop={handleDrop}
        onDragEnd={() => {
          if (!droppedRef.current) {
            resetDragState();
          }
          window.setTimeout(() => {
            setIgnoreClick(false);
          }, 0);
        }}
        onStateChange={handleStateChange}
      />
    );
  };

  return (
    <div className={slots.root()}>
      <div className={slots.tools()}>
        <div className={slots.search()}>
          <TextField
            icon={<MagnifyingGlassIcon />}
            value={queryInput}
            onValueChange={setQueryInput}
            placeholder={t("tasksCard.searchPlaceholder")}
            aria-label={t("tasksCard.searchPlaceholder")}
          />
        </div>
        <div className={slots.filters()}>
          <div className={slots.filter()}>
            <Select
              value={selectedState}
              onValueChange={setState}
            >
              <SelectTrigger
                size={2}
                placeholder={t("tasksCard.filters.allStatuses")}
                aria-label={t("tasksCard.filters.status")}
              >
                {(value: TaskState | null) => value == null
                  ? t("tasksCard.filters.allStatuses")
                  : (
                      <span className={slots.stateOption()}>
                        <TaskStateIcon state={value} />
                        {t(`tasksCard.states.${taskStateKeys[value]}`)}
                      </span>
                    )}
              </SelectTrigger>
              <SelectPopup align="start">
                <SelectItem value={null}>
                  {t("tasksCard.filters.allStatuses")}
                </SelectItem>
                {taskStates.map(state => (
                  <SelectItem key={state} value={state}>
                    <span className={slots.stateOption()}>
                      <TaskStateIcon state={state} />
                      {t(`tasksCard.states.${taskStateKeys[state]}`)}
                    </span>
                  </SelectItem>
                ))}
              </SelectPopup>
            </Select>
          </div>
        </div>
      </div>
      {filteredTasks.length === 0
        ? (
            <Text size={2} color="faint" align="center" className={slots.empty()}>
              {query !== "" || selectedState != null
                ? t("tasksCard.noResults")
                : t("tasksCard.empty")}
            </Text>
          )
        : (
            <Card padding="none">
              <TooltipProvider>
                <div className={slots.list()}>
                  {selectedState == null
                    ? taskStates
                        .filter(state =>
                          tasksPerState.get(state)?.length
                          || (draggedId && dropTargetState === state),
                        )
                        .map((state) => {
                          const stateTasks = tasksPerState.get(state) ?? [];
                          const displayEdges = applyPreviewOrder(stateTasks);
                          return (
                            <Fragment key={state}>
                              <div
                                className={slots.sectionHeader()}
                                onDragOver={canDrag
                                  ? (e) => {
                                      e.preventDefault();
                                      setDropTargetState(state);
                                    }
                                  : undefined}
                                onDrop={canDrag ? handleDrop : undefined}
                              >
                                <TaskStateIcon state={state} />
                                <Text size={2} weight="medium">
                                  {t(`tasksCard.states.${taskStateKeys[state]}`)}
                                </Text>
                                <Badge>{stateTasks.length}</Badge>
                              </div>
                              {displayEdges.map(({ node }) =>
                                renderTaskRow(node, state))}
                            </Fragment>
                          );
                        })
                    : displayTasks.map(({ node }) => renderTaskRow(node))}
                </div>
              </TooltipProvider>
            </Card>
          )}
      {canDrag && filteredTasks.length > 1 && (
        <Text size={2} color="faint">
          {selectedState == null
            ? t("tasksCard.dragInstructions.all")
            : t("tasksCard.dragInstructions.state")}
        </Text>
      )}
    </div>
  );
}
