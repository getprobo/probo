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

export {
  emptyDeadlineBounds,
  type TaskDeadlineBounds,
  type TaskDueFilter,
  taskDueFilterBounds,
  taskDueFilterFromParam,
  taskDueFilters,
  taskMatchesDeadline,
} from "./taskDueFilter";
export { registerTasksCatalogs, TASKS_NAMESPACE } from "./i18n";
export { isOverdueTask } from "./isOverdueTask";
export { TaskDetails, type TaskDetailsProps } from "./TaskDetails";
export {
  TaskActivityEmpty,
  TaskCommentsEmpty,
  TaskEngagement,
  type TaskEngagementProps,
} from "./TaskEngagement";
export {
  flattenTaskListItems,
  type TaskListFlatItem,
  type TaskListGroupInput,
  type TaskListRowPosition,
  taskListRowPosition,
} from "./flattenTaskListItems";
export {
  TaskListGroupHeader,
  type TaskListGroupHeaderProps,
} from "./TaskListGroupHeader";
export {
  TaskListDueDateFilter,
  type TaskListDueDateFilterProps,
  TaskListSearchField,
  type TaskListSearchFieldProps,
  TaskListStatusFilter,
  type TaskListStatusFilterProps,
  TaskListToolbar,
  type TaskListToolbarProps,
} from "./TaskListToolbar";
export {
  TaskListView,
  type TaskListViewGroup,
  type TaskListViewItem,
  type TaskListViewProps,
} from "./TaskListView";
export {
  TaskListRowFrame,
  type TaskListRowFrameProps,
} from "./TaskListRowFrame";
export { taskListPageSize } from "./taskListPageSize";
export {
  type TaskListFillGroup,
  type TaskListLoadNext,
  type TaskListLoadNextOptions,
  type TasksListFillRange,
  useTasksListFill,
  type UseTasksListFillOptions,
} from "./useTasksListFill";
export {
  taskListGroupHeaderGap,
  taskListGroupHeaderOffset,
  taskListGroupHeaderSize,
  taskListItemSize,
  taskListRowSize,
} from "./taskListSize";
export { TaskNameField, type TaskNameFieldProps } from "./TaskNameField";
export {
  PropertyRow,
  TaskProperties,
  type TaskPropertiesChange,
  type TaskPropertiesProps,
  type TaskPropertiesTask,
} from "./TaskProperties";
export {
  TaskDurationField,
  type TaskDurationFieldProps,
} from "./TaskDurationField";
export {
  type ParsedTaskDuration,
  parseTaskDuration,
  parseTaskDurationAmount,
  stringifyTaskDuration,
  taskDurationsEqual,
  type TaskDurationUnit,
  taskDurationUnits,
  taskEstimateDurationUnits,
  taskRecurrenceDurationUnits,
} from "./taskDuration";
export {
  formatTaskActivityValue,
  type TaskActivityField,
  taskActivitySentenceKey,
  type TaskActivityType,
} from "./taskActivity";
export {
  taskCommentEditor,
  taskCommentForm,
  taskCommentListItem,
  taskCommentsSection,
  taskDescription,
  taskDetails,
  taskDurationField,
  taskEngagement,
  taskList,
  taskListItem,
  taskListRow,
  taskListToolbar,
  taskListView,
  taskNameField,
  taskProperties,
} from "./variants";
export {
  TaskListItem,
  type TaskListItemAssignee,
  type TaskListItemProps,
  type TaskListItemTask,
} from "./TaskListItem";
export { TaskPriorityIcon } from "./TaskPriorityIcon";
export { TaskStateIcon } from "./TaskStateIcon";
export {
  pendingTaskStates,
  taskListGroupStates,
  taskPriorities,
  type TaskPriority,
  taskPriorityKeys,
  type TaskState,
  taskStateKeys,
  taskStates,
} from "./taskState";
