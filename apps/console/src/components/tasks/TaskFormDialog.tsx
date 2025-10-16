import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DurationPicker,
  Input,
  Label,
  PropertyRow,
  Textarea,
  useDialogRef,
  type DialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslate } from "@probo/i18n";
import { Breadcrumb } from "@probo/ui";
import { graphql } from "relay-runtime";
import { useFragment, useRelayEnvironment } from "react-relay";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import type { TaskFormDialogFragment$key } from "./__generated__/TaskFormDialogFragment.graphql";
import { MeasureSelectField } from "/components/form/MeasureSelectField";
import { Controller } from "react-hook-form";
import { updateStoreCounter } from "/hooks/useMutationWithIncrement";
import { formatDatetime } from "@probo/helpers";

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

const schema = z.object({
  name: z.string(),
  description: z.string(),
  timeEstimate: z.string().nullable(),
  assignedToId: z.string(),
  measureId: z.string(),
  deadline: z.string().optional(),
});

type Props = {
  children?: ReactNode;
  task?: TaskFormDialogFragment$key;
  connection?: string;
  ref?: DialogRef;
  measureId?: string;
};

export default function TaskFormDialog(props: Props) {
  const { __ } = useTranslate();
  const dialogRef = props.ref ?? useDialogRef();
  const organizationId = useOrganizationId();
  const task = useFragment(taskFragment, props.task);
  const relayEnv = useRelayEnvironment();
  const [mutate] = task
    ? useMutationWithToasts(taskUpdateMutation, {
        successMessage: __("Task updated successfully."),
        errorMessage: __("Failed to update task. Please try again."),
      })
    : useMutationWithToasts(taskCreateMutation, {
        successMessage: __("Task created successfully."),
        errorMessage: __("Failed to create task. Please try again."),
      });

  const { control, handleSubmit, register, formState, reset } =
    useFormWithSchema(schema, {
      defaultValues: {
        name: task?.name ?? "",
        description: task?.description ?? "",
        timeEstimate: task?.timeEstimate ?? "",
        assignedToId: task?.assignedTo?.id ?? "",
        measureId: task?.measure?.id ?? props.measureId ?? "",
        deadline: task?.deadline?.split("T")[0] ?? "",
      },
    });

  const onSubmit = handleSubmit(async (data) => {
    if (task) {
      await mutate({
        variables: {
          input: {
            taskId: task.id,
            name: data.name,
            description: data.description,
            timeEstimate: data.timeEstimate || null,
            deadline: formatDatetime(data.deadline) ?? null,
          },
        },
      });
    } else {
      await mutate({
        variables: {
          input: {
            organizationId,
            name: data.name,
            description: data.description,
            timeEstimate: data.timeEstimate || null,
            deadline: formatDatetime(data.deadline) ?? null,
            assignedToId: data.assignedToId,
            measureId: data.measureId,
          },
          connections: [props.connection!],
        },
        onCompleted: (_response, errors) => {
          if (!errors) {
            updateStoreCounter(relayEnv, data.measureId, "tasks(first:0)", 1);
          }
        },
      });
      reset();
    }
    dialogRef.current?.close();
  });
  const isUpdating = !!task;
  const showMeasure = !props.measureId && !isUpdating;
  const isCreating = !isUpdating;

  return (
    <Dialog
      ref={dialogRef}
      trigger={props.children}
      title={
        <Breadcrumb
          items={[__("Tasks"), isUpdating ? __("Edit Task") : __("New Task")]}
        />
      }
    >
      <form onSubmit={onSubmit}>
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
            {isCreating && (
              <PropertyRow
                label={__("Assigned to")}
                error={formState.errors.assignedToId?.message}
              >
                <PeopleSelectField
                  name="assignedToId"
                  control={control}
                  organizationId={organizationId}
                />
              </PropertyRow>
            )}
            {showMeasure && (
              <PropertyRow
                label={__("Measure")}
                error={formState.errors.measureId?.message}
              >
                <MeasureSelectField
                  name="measureId"
                  control={control}
                  organizationId={organizationId}
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
                render={({ field: { onChange, ...field } }) => (
                  <DurationPicker
                    {...field}
                    onValueChange={(value) => onChange(value)}
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
