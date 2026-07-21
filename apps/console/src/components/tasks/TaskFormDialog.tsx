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

import { formatDatetime } from "@probo/helpers";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  type DialogRef,
  DurationPicker,
  Input,
  Label,
  Option,
  PriorityLevel,
  PropertyRow,
  Select,
  TaskStateIcon,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { Breadcrumb } from "@probo/ui";
import { type ReactNode, useEffect } from "react";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useFragment, useRelayEnvironment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { TaskFormDialogFragment$key } from "#/__generated__/core/TaskFormDialogFragment.graphql";
import { MeasureSelectField } from "#/components/form/MeasureSelectField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const taskFragment = graphql`
  fragment TaskFormDialogFragment on Task {
    id
    description
    name
    state
    priority
    timeEstimate
    deadline
    assignedTo {
      id
    }
    measure {
      id
    }
  }
`;

const taskCreateMutation = graphql`
  mutation TaskFormDialogCreateMutation(
    $input: CreateTaskInput!
    $connections: [ID!]!
  ) {
    createTask(input: $input) {
      taskEdge @appendEdge(connections: $connections) {
        node {
          ...TaskFormDialogFragment
          ...TasksCard_task
          ...TasksCard_TaskRowFragment
        }
      }
    }
  }
`;

export const taskUpdateMutation = graphql`
  mutation TaskFormDialogUpdateMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TaskFormDialogFragment
        ...TasksCard_task
        ...TasksCard_TaskRowFragment
      }
    }
  }
`;

export const taskStates = ["TODO", "IN_PROGRESS", "DONE"] as const;
export const taskPriorities = ["URGENT", "HIGH", "MEDIUM", "LOW"] as const;

const createTaskSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional().nullable(),
  priority: z.enum(taskPriorities),
  timeEstimate: z.string().optional().nullable(),
  assignedToId: z.string().optional().nullable(),
  measureId: z.preprocess(
    val => (val === "" || val == null ? null : val),
    z.string().nullable().optional(),
  ),
  deadline: z.string().optional().nullable(),
});

const updateTaskSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional().nullable(),
  state: z.enum(taskStates),
  priority: z.enum(taskPriorities),
  timeEstimate: z.string().optional().nullable(),
  assignedToId: z.preprocess(
    val => (val === "" || val == null ? null : val),
    z.string().nullable().optional(),
  ),
  measureId: z.preprocess(
    val => (val === "" || val == null ? null : val),
    z.string().nullable().optional(),
  ),
  deadline: z.string().optional().nullable(),
});

type Props = {
  children?: ReactNode;
  task?: TaskFormDialogFragment$key;
  connection?: string;
  ref?: DialogRef;
  measureId?: string;
  onCompleted?: () => void;
};

export default function TaskFormDialog(props: Props) {
  const { children, connection, ref, task: taskKey, measureId, onCompleted } = props;
  const { t } = useTranslation();
  const newRef = useDialogRef();
  const dialogRef = ref ?? newRef;
  const organizationId = useOrganizationId();
  const task = useFragment(taskFragment, taskKey);
  const relayEnv = useRelayEnvironment();
  const [mutate] = useMutationWithToasts(
    task ? taskUpdateMutation : taskCreateMutation,
    {
      successMessage: t(task ? "taskFormDialog.messages.updated" : "taskFormDialog.messages.created"),
      errorMessage: t(task ? "taskFormDialog.errors.update" : "taskFormDialog.errors.create"),
    },
  );

  const isUpdating = !!task;

  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(isUpdating ? updateTaskSchema : createTaskSchema, {
      defaultValues: {
        name: task?.name ?? "",
        description: task?.description ?? "",
        state: task?.state ?? "TODO",
        priority: task?.priority ?? "MEDIUM",
        timeEstimate: task?.timeEstimate ?? "",
        assignedToId: task?.assignedTo?.id ?? "",
        measureId: task?.measure?.id ?? measureId ?? "",
        deadline: task?.deadline?.split("T")[0] ?? "",
      },
    });

  useEffect(() => {
    if (task) {
      reset({
        name: task.name,
        description: task.description ?? "",
        state: task.state,
        priority: task.priority,
        timeEstimate: task.timeEstimate ?? "",
        assignedToId: task.assignedTo?.id ?? "",
        measureId: task.measure?.id ?? measureId ?? "",
        deadline: task.deadline?.split("T")[0] ?? "",
      });
    }
  }, [
    task, reset, measureId,
  ]);

  const onSubmit = async (data: z.infer<typeof updateTaskSchema | typeof createTaskSchema>) => {
    if (task) {
      await mutate({
        variables: {
          input: {
            taskId: task.id,
            name: data.name,
            description: data.description || null,
            state: "state" in data ? data.state : undefined,
            priority: data.priority,
            timeEstimate: data.timeEstimate || null,
            deadline: formatDatetime(data.deadline) ?? null,
            assignedToId: data.assignedToId ?? null,
            measureId: data.measureId || null,
          },
        },
        onCompleted: (_response, errors) => {
          if (!errors) onCompleted?.();
        },
      });
    } else {
      await mutate({
        variables: {
          input: {
            organizationId,
            name: data.name,
            description: data.description || null,
            priority: data.priority,
            timeEstimate: data.timeEstimate || null,
            deadline: formatDatetime(data.deadline) ?? null,
            assignedToId: data.assignedToId || null,
            measureId: data.measureId || null,
          },
          connections: [connection!],
        },
        onCompleted: (_response, errors) => {
          if (!errors) {
            if (data.measureId) {
              updateStoreCounter(relayEnv, data.measureId, "tasks(first:0)", 1);
            }
            onCompleted?.();
          }
        },
      });
      reset();
    }
    dialogRef.current?.close();
  };
  const showMeasure = !measureId;

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={(
        <Breadcrumb
          items={[t("taskFormDialog.breadcrumb.tasks"), isUpdating ? t("taskFormDialog.breadcrumb.edit") : t("taskFormDialog.breadcrumb.new")]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          <div className="py-8 px-10 space-y-4">
            <Input
              id="title"
              required
              variant="title"
              placeholder={t("taskFormDialog.fields.title.placeholder")}
              {...register("name")}
            />
            <Textarea
              id="content"
              variant="ghost"
              autogrow
              placeholder={t("taskFormDialog.fields.description.placeholder")}
              {...register("description")}
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{t("taskFormDialog.properties")}</Label>
            {isUpdating && (
              <PropertyRow
                label={t("taskFormDialog.fields.state.label")}
                error={"state" in formState.errors ? formState.errors.state?.message : undefined}
              >
                <Controller
                  name="state"
                  control={control}
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      <Option value="TODO">
                        <span className="flex items-center gap-2">
                          <TaskStateIcon state="TODO" />
                          {t("taskFormDialog.states.todo")}
                        </span>
                      </Option>
                      <Option value="IN_PROGRESS">
                        <span className="flex items-center gap-2">
                          <TaskStateIcon state="IN_PROGRESS" />
                          {t("taskFormDialog.states.inProgress")}
                        </span>
                      </Option>
                      <Option value="DONE">
                        <span className="flex items-center gap-2">
                          <TaskStateIcon state="DONE" />
                          {t("taskFormDialog.states.done")}
                        </span>
                      </Option>
                    </Select>
                  )}
                />
              </PropertyRow>
            )}
            <PropertyRow
              label={t("taskFormDialog.fields.priority.label")}
              error={formState.errors.priority?.message}
            >
              <Controller
                name="priority"
                control={control}
                render={({ field }) => (
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    <Option value="URGENT">
                      <span className="flex items-center gap-2">
                        <PriorityLevel level="URGENT" />
                        {t("taskFormDialog.priorities.urgent")}
                      </span>
                    </Option>
                    <Option value="HIGH">
                      <span className="flex items-center gap-2">
                        <PriorityLevel level="HIGH" />
                        {t("taskFormDialog.priorities.high")}
                      </span>
                    </Option>
                    <Option value="MEDIUM">
                      <span className="flex items-center gap-2">
                        <PriorityLevel level="MEDIUM" />
                        {t("taskFormDialog.priorities.medium")}
                      </span>
                    </Option>
                    <Option value="LOW">
                      <span className="flex items-center gap-2">
                        <PriorityLevel level="LOW" />
                        {t("taskFormDialog.priorities.low")}
                      </span>
                    </Option>
                  </Select>
                )}
              />
            </PropertyRow>
            <PropertyRow
              label={t("taskFormDialog.fields.assignedTo.label")}
              error={formState.errors.assignedToId?.message}
            >
              <PeopleSelectField
                name="assignedToId"
                control={control}
                organizationId={organizationId}
                optional={true}
              />
            </PropertyRow>
            {showMeasure && (
              <PropertyRow
                label={t("taskFormDialog.fields.measure.label")}
                error={formState.errors.measureId?.message}
              >
                <MeasureSelectField
                  name="measureId"
                  control={control}
                  organizationId={organizationId}
                  optional={true}
                />
              </PropertyRow>
            )}
            <PropertyRow
              label={t("taskFormDialog.fields.timeEstimate.label")}
              error={formState.errors.timeEstimate?.message}
            >
              <Controller
                name="timeEstimate"
                control={control}
                render={({ field: { onChange, value, ...field } }) => (
                  <DurationPicker
                    {...field}
                    value={value ?? null}
                    onValueChange={value => onChange(value)}
                  />
                )}
              />
            </PropertyRow>
            <PropertyRow
              label={t("taskFormDialog.fields.deadline.label")}
              error={formState.errors.deadline?.message}
            >
              <Input id="deadline" type="date" {...register("deadline")} />
            </PropertyRow>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit">
            {isUpdating ? t("taskFormDialog.actions.update") : t("taskFormDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
