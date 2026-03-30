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

import { formatDatetime } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { Breadcrumb } from "@probo/ui";
import type { ReactNode } from "react";
import { Controller } from "react-hook-form";
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
  const { __ } = useTranslate();
  const newRef = useDialogRef();
  const dialogRef = ref ?? newRef;
  const organizationId = useOrganizationId();
  const task = useFragment(taskFragment, taskKey);
  const relayEnv = useRelayEnvironment();
  const [mutate] = useMutationWithToasts(
    task ? taskUpdateMutation : taskCreateMutation,
    {
      successMessage: __(`Task ${task ? "updated" : "created"} successfully.`),
      errorMessage: __(`Failed to ${task ? "update" : "create"} task`),
    },
  );

  const isUpdating = !!task;

  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(isUpdating ? updateTaskSchema : createTaskSchema, {
      defaultValues: {
        name: task?.name ?? "",
        description: task?.description ?? "",
        priority: task?.priority ?? "MEDIUM",
        timeEstimate: task?.timeEstimate ?? "",
        assignedToId: task?.assignedTo?.id ?? "",
        measureId: task?.measure?.id ?? measureId ?? "",
        deadline: task?.deadline?.split("T")[0] ?? "",
      },
    });

  const onSubmit = async (data: z.infer<typeof updateTaskSchema | typeof createTaskSchema>) => {
    if (task) {
      await mutate({
        variables: {
          input: {
            taskId: task.id,
            name: data.name,
            description: data.description || null,
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
          items={[__("Tasks"), isUpdating ? __("Edit Task") : __("New Task")]}
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
              placeholder={__("Task title")}
              {...register("name")}
            />
            <Textarea
              id="content"
              variant="ghost"
              autogrow
              placeholder={__("Add description")}
              {...register("description")}
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{__("Properties")}</Label>
            <PropertyRow
              label={__("Priority")}
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
                        {__("Urgent")}
                      </span>
                    </Option>
                    <Option value="HIGH">
                      <span className="flex items-center gap-2">
                        <PriorityLevel level="HIGH" />
                        {__("High")}
                      </span>
                    </Option>
                    <Option value="MEDIUM">
                      <span className="flex items-center gap-2">
                        <PriorityLevel level="MEDIUM" />
                        {__("Medium")}
                      </span>
                    </Option>
                    <Option value="LOW">
                      <span className="flex items-center gap-2">
                        <PriorityLevel level="LOW" />
                        {__("Low")}
                      </span>
                    </Option>
                  </Select>
                )}
              />
            </PropertyRow>
            <PropertyRow
              label={__("Assigned to")}
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
                label={__("Measure")}
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
              label={__("Time estimate")}
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
              label={__("Deadline")}
              error={formState.errors.deadline?.message}
            >
              <Input id="deadline" type="date" {...register("deadline")} />
            </PropertyRow>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit">
            {isUpdating ? __("Update task") : __("Create task")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
