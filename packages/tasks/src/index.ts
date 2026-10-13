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

export { registerTasksCatalogs, TASKS_NAMESPACE } from "./i18n";
export { isOverdueTask } from "./isOverdueTask";
export { TaskDetails, type TaskDetailsProps } from "./TaskDetails";
export {
  TaskActivityEmpty,
  TaskCommentsEmpty,
  TaskEngagement,
  type TaskEngagementProps,
} from "./TaskEngagement";
export { TaskList, type TaskListItemSource, type TaskListProps } from "./TaskList";
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
  parseTaskDuration,
  parseTaskDurationAmount,
  stringifyTaskDuration,
  taskDurationsEqual,
  taskDurationUnits,
  taskEstimateDurationUnits,
  taskRecurrenceDurationUnits,
  type ParsedTaskDuration,
  type TaskDurationUnit,
} from "./taskDuration";
export {
  formatTaskActivityValue,
  taskActivitySentenceKey,
  type TaskActivityField,
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
  type TaskPriority,
  taskPriorities,
  taskPriorityKeys,
  type TaskState,
  taskStateKeys,
  taskStates,
} from "./taskState";
