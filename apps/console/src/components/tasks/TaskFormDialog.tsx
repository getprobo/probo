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
  PropertyRow,
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
    state
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
      taskEdge @prependEdge(connections: $connections) {
        node {
          ...TaskFormDialogFragment
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
      }
    }
  }
`;

const createTaskSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional().nullable(),
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
};

export default function TaskFormDialog(props: Props) {
  const { children, connection, ref, task: taskKey, measureId } = props;
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
            timeEstimate: data.timeEstimate || null,
            deadline: formatDatetime(data.deadline) ?? null,
            assignedToId: data.assignedToId ?? null,
            measureId: data.measureId || null,
          },
        },
      });
    } else {
      await mutate({
        variables: {
          input: {
            organizationId,
            name: data.name,
            description: data.description || null,
            timeEstimate: data.timeEstimate || null,
            deadline: formatDatetime(data.deadline) ?? null,
            assignedToId: data.assignedToId || null,
            measureId: data.measureId || null,
          },
          connections: [connection!],
        },
        onCompleted: (_response, errors) => {
          if (!errors && data.measureId) {
            updateStoreCounter(relayEnv, data.measureId, "tasks(first:0)", 1);
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
