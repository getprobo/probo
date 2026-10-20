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

import { UserCircleDashedIcon } from "@phosphor-icons/react";
import { TaskListItem as SharedTaskListItem, type TaskPriority, type TaskState } from "@probo/tasks";
import { memo, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useRelayEnvironment } from "react-relay";
import type { RecordSourceSelectorProxy } from "relay-runtime";

import type { TaskListItem_task$key } from "#/__generated__/core/TaskListItem_task.graphql";
import type { TaskListItemUpdateStateMutation } from "#/__generated__/core/TaskListItemUpdateStateMutation.graphql";
import { TaskAssigneeMenu } from "#/components/tasks/TaskAssigneeMenu";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { TaskDeleteDialog } from "#/pages/organizations/tasks/_components/TaskDeleteDialog";
import { TaskListLinearDialog } from "#/pages/organizations/tasks/_components/TaskListLinearDialog";
import { TaskListLinearMenu } from "#/pages/organizations/tasks/_components/TaskListLinearMenu";
import {
  reconcileOrganizationTaskListMembership,
  resortOrganizationTaskListEdge,
  taskListStateMoveUpdaters,
} from "#/pages/organizations/tasks/_lib/moveOrganizationTaskListEdge";
import {
  organizationTaskListFilterVariants,
  taskDetailsPath,
} from "#/pages/organizations/tasks/_lib/taskPath";
import { useTasksCardFilters } from "#/pages/organizations/tasks/_lib/useTasksCardFilters";

const noneValue = "__NONE__";

const taskListItemFragment = graphql`
  fragment TaskListItem_task on Task {
    id
    name
    state
    priority
    createdAt
    timeEstimate
    deadline
    recurrenceInterval
    canUpdate: permission(action: "core:task:update")
    canDelete: permission(action: "core:task:delete")
    externalLink {
      identifier
      url
    }
    assignedTo {
      id
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
    ...TaskDeleteDialog_task
    ...TaskListLinearDialog_task
    ...TaskListLinearMenu_task
  }
`;

const updateStateMutation = graphql`
  mutation TaskListItemUpdateStateMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TasksPageList_task
        ...TaskListItem_task
        ...TaskDetailsPage_task
      }
    }
  }
`;

interface TaskListItemProps {
  taskKey: TaskListItem_task$key;
  sectionState?: TaskState;
  canDrag?: boolean;
  isDragging?: boolean;
  isGhost?: boolean;
  ignoreClick: boolean;
  linearConnected?: boolean;
  onDragStart?: () => void;
  onDragOver?: (event: React.DragEvent) => void;
  onDrop?: () => void;
  onDragEnd?: () => void;
  onStateChange?: () => void;
}

function taskListItemPropsAreEqual(
  previous: TaskListItemProps,
  next: TaskListItemProps,
): boolean {
  return previous.taskKey === next.taskKey
    && previous.sectionState === next.sectionState
    && previous.canDrag === next.canDrag
    && previous.isDragging === next.isDragging
    && previous.isGhost === next.isGhost
    && previous.ignoreClick === next.ignoreClick
    && previous.linearConnected === next.linearConnected;
}

export const TaskListItem = memo(function TaskListItem(props: TaskListItemProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const environment = useRelayEnvironment();
  const { graphqlFilter } = useTasksCardFilters();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [linearOpen, setLinearOpen] = useState(false);
  const task = useFragment(taskListItemFragment, props.taskKey);
  const { canUpdate, canDelete } = task;
  const [updateState, isUpdating] = useMutation<TaskListItemUpdateStateMutation>(
    updateStateMutation,
  );
  const assignedTo = task.assignedTo;

  function updateTask(input: {
    state?: TaskState;
    assignedToId?: string | null;
    priority?: TaskPriority;
    deadline?: string | null;
    timeEstimate?: string | null;
    recurrenceInterval?: string | null;
  }) {
    const listFilter = {
      query: graphqlFilter.query,
      assignedToId: graphqlFilter.assignedToId,
      internalControlId: graphqlFilter.internalControlId,
      deadlineOnOrBefore: graphqlFilter.deadlineOnOrBefore,
      deadlineOnOrAfter: graphqlFilter.deadlineOnOrAfter,
      hasDeadline: graphqlFilter.hasDeadline,
      hasAssignee: graphqlFilter.hasAssignee,
    };
    const filters = organizationTaskListFilterVariants(listFilter);
    const previousAssignedToId = task.assignedTo?.id ?? null;
    const assigneeChanged = "assignedToId" in input;
    const move = input.state == null
      ? null
      : taskListStateMoveUpdaters(environment, {
          taskId: task.id,
          nextState: input.state,
          organizationId,
          filter: listFilter,
        });

    function linkAssignee(store: RecordSourceSelectorProxy) {
      if (!assigneeChanged) {
        return;
      }
      const node = store.get(task.id);
      if (node == null) {
        return;
      }
      const assignedToId = input.assignedToId ?? null;
      if (assignedToId == null) {
        node.setValue(null, "assignedTo");
        return;
      }
      const profile = store.get(assignedToId);
      if (profile != null) {
        node.setLinkedRecord(profile, "assignedTo");
      }
    }

    function reconcileAssignee(store: RecordSourceSelectorProxy) {
      if (!assigneeChanged) {
        return;
      }
      const node = store.get(task.id);
      if (node == null) {
        return;
      }
      reconcileOrganizationTaskListMembership(store, node, {
        organizationId,
        state: input.state ?? task.state,
        filters,
        previousAssignedToId,
        previousInternalControlIds: null,
      });
    }

    void updateState({
      variables: {
        input: {
          taskId: task.id,
          ...input,
        },
      },
      optimisticUpdater(store) {
        move?.optimisticUpdater(store);
        linkAssignee(store);
        reconcileAssignee(store);
        if (input.state == null && input.priority != null) {
          const node = store.get(task.id);
          if (node != null) {
            node.setValue(input.priority, "priority");
            resortOrganizationTaskListEdge(
              store,
              node,
              organizationId,
              task.state,
              filters,
            );
          }
        }
      },
      updater(store) {
        move?.updater(store);
        reconcileAssignee(store);
        if (input.state == null && input.priority != null) {
          const node = store.get(task.id);
          if (node != null) {
            resortOrganizationTaskListEdge(
              store,
              node,
              organizationId,
              task.state,
              filters,
            );
          }
        }
      },
      onCompleted: () => {
        if (input.state != null) {
          props.onStateChange?.();
        }
      },
    }).catch(() => {
      // Error feedback is handled by useMutation.
    });
  }

  const assigneeSelectProps = {
    value: assignedTo?.id ?? null,
    disabled: isUpdating || props.ignoreClick,
    onValueChange: (assignedToId: string | null) => {
      updateTask({ assignedToId });
    },
    emptyLabel: t("detailsPage.unassigned"),
    emptyValue: noneValue,
    ariaLabel: t("detailsPage.fields.assignedTo"),
    size: 1 as const,
    contractEnded: false,
    pinned: assignedTo == null
      ? null
      : {
          id: assignedTo.id,
          fullName: assignedTo.fullName,
          emailAddress: assignedTo.emailAddress,
          avatarUrl: assignedTo.avatar?.downloadUrl,
        },
  };

  return (
    <>
      <SharedTaskListItem
        task={{
          id: task.id,
          name: task.name,
          state: task.state,
          priority: task.priority,
          createdAt: task.createdAt,
          timeEstimate: task.timeEstimate,
          deadline: task.deadline,
          recurrenceInterval: task.recurrenceInterval,
          externalLink: task.externalLink == null
            ? null
            : {
                identifier: task.externalLink.identifier,
                url: task.externalLink.url,
              },
          assignedTo: assignedTo == null
            ? null
            : {
                id: assignedTo.id,
                fullName: assignedTo.fullName,
                emailAddress: assignedTo.emailAddress,
                avatarUrl: assignedTo.avatar?.downloadUrl,
              },
        }}
        canUpdate={canUpdate}
        canUpdateAssignee={canUpdate}
        isUpdating={isUpdating}
        detailsTo={taskDetailsPath(organizationId, task.id)}
        renderAssigneeMenu={canUpdate
          ? () => (
              <TaskAssigneeMenu
                value={assigneeSelectProps.value}
                emptyLabel={assigneeSelectProps.emptyLabel}
                emptyValue={assigneeSelectProps.emptyValue}
                emptyIcon={<UserCircleDashedIcon />}
                disabled={assigneeSelectProps.disabled}
                onValueChange={assigneeSelectProps.onValueChange}
              />
            )
          : undefined}
        onDelete={canDelete
          ? () => {
              setDeleteOpen(true);
            }
          : undefined}
        extraMenu={props.linearConnected && canUpdate
          ? (
              <TaskListLinearMenu
                taskKey={task}
                linked={task.externalLink != null}
                onLink={() => {
                  setLinearOpen(true);
                }}
              />
            )
          : undefined}
        sectionState={props.sectionState}
        canDrag={props.canDrag}
        isDragging={props.isDragging}
        isGhost={props.isGhost}
        ignoreClick={props.ignoreClick}
        onStateChange={(state) => {
          updateTask({ state });
        }}
        onPriorityChange={(priority) => {
          updateTask({ priority });
        }}
        onDeadlineChange={(deadline) => {
          updateTask({
            deadline,
            ...(deadline == null && task.recurrenceInterval != null
              ? { recurrenceInterval: null }
              : {}),
          });
        }}
        onTimeEstimateChange={(timeEstimate) => {
          updateTask({ timeEstimate });
        }}
        onDragStart={props.onDragStart}
        onDragOver={props.onDragOver}
        onDrop={props.onDrop}
        onDragEnd={props.onDragEnd}
      />
      {props.linearConnected && canUpdate && task.externalLink == null && (
        <TaskListLinearDialog
          taskKey={task}
          open={linearOpen}
          onOpenChange={setLinearOpen}
        />
      )}
      {canDelete && (
        <TaskDeleteDialog
          taskKey={task}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
        />
      )}
    </>
  );
}, taskListItemPropsAreEqual);
