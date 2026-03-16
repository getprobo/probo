import {
  formatDatetime,
  formatError,
  getStatusOptions,
  type GraphQLError,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Option,
  Select,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode } from "react";
import { Controller } from "react-hook-form";
import { graphql, useMutation } from "react-relay";
import { z } from "zod";

import type { CreateFindingDialogMutation } from "#/__generated__/core/CreateFindingDialogMutation.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const createFindingMutation = graphql`
  mutation CreateFindingDialogMutation(
    $input: CreateFindingInput!
    $connections: [ID!]!
  ) {
    createFinding(input: $input) {
      findingEdge @prependEdge(connections: $connections) {
        node {
          id
          kind
          referenceId
          description
          source
          identifiedOn
          rootCause
          correctiveAction
          dueDate
          status
          priority
          effectivenessCheck
          owner {
            id
            fullName
          }
          createdAt
          canUpdate: permission(action: "core:finding:update")
          canDelete: permission(action: "core:finding:delete")
        }
      }
    }
  }
`;

const schema = z.object({
  kind: z.enum(["NONCONFORMITY", "OBSERVATION", "EXCEPTION"]),
  description: z.string().optional(),
  source: z.string().optional(),
  identifiedOn: z.string().optional(),
  rootCause: z.string().optional(),
  correctiveAction: z.string().optional(),
  ownerId: z.string().nullable().optional(),
  dueDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED", "MITIGATED", "FALSE_POSITIVE"]),
  priority: z.enum(["LOW", "MEDIUM", "HIGH"]),
  effectivenessCheck: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

interface CreateFindingDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionIds?: string[];
}

export function CreateFindingDialog({
  children,
  organizationId,
  connectionIds,
}: CreateFindingDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [createFinding] = useMutation<CreateFindingDialogMutation>(createFindingMutation);
  const statusOptions = getStatusOptions(__).filter(
    opt => opt.value !== "RISK_ACCEPTED",
  );

  const kindOptions = [
    { value: "NONCONFORMITY", label: __("Nonconformity") },
    { value: "OBSERVATION", label: __("Observation") },
    { value: "EXCEPTION", label: __("Exception") },
  ];

  const priorityOptions = [
    { value: "LOW", label: __("Low") },
    { value: "MEDIUM", label: __("Medium") },
    { value: "HIGH", label: __("High") },
  ];

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      kind: "NONCONFORMITY" as const,
      description: "",
      source: "",
      identifiedOn: "",
      rootCause: "",
      correctiveAction: "",
      ownerId: null,
      dueDate: "",
      status: "OPEN" as const,
      priority: "MEDIUM" as const,
      effectivenessCheck: "",
    },
  });

  const onSubmit = (formData: FormData) => {
    createFinding({
      variables: {
        input: {
          organizationId,
          kind: formData.kind,
          description: formData.description || undefined,
          source: formData.source || undefined,
          identifiedOn: formatDatetime(formData.identifiedOn),
          rootCause: formData.rootCause || undefined,
          correctiveAction: formData.correctiveAction || undefined,
          ownerId: formData.ownerId || undefined,
          dueDate: formatDatetime(formData.dueDate),
          status: formData.status,
          priority: formData.priority,
          effectivenessCheck: formData.effectivenessCheck || undefined,
        },
        connections: connectionIds ?? [],
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Finding created successfully"),
          variant: "success",
        });
        reset();
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to create finding"), error as GraphQLError),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Findings"), __("Create")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Controller
            control={control}
            name="kind"
            render={({ field }) => (
              <Field label={__("Kind")} required>
                <Select
                  variant="editor"
                  placeholder={__("Select kind")}
                  onValueChange={field.onChange}
                  value={field.value}
                  className="w-full"
                >
                  {kindOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      {option.label}
                    </Option>
                  ))}
                </Select>
                {formState.errors.kind && (
                  <p className="text-sm text-red-500 mt-1">{formState.errors.kind.message}</p>
                )}
              </Field>
            )}
          />

          <div className="space-y-2">
            <Label htmlFor="description">{__("Description")}</Label>
            <Textarea
              id="description"
              {...register("description")}
              placeholder={__("Brief description of the finding...")}
              rows={2}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Field
              label={__("Source")}
              {...register("source")}
              placeholder={__("Enter source")}
              error={formState.errors.source?.message}
            />

            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={__("Owner")}
              error={formState.errors.ownerId?.message}
              optional
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Field label={__("Status")}>
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <Select
                    variant="editor"
                    placeholder={__("Select status")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                  >
                    {statusOptions.map(option => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                )}
              />
              {formState.errors.status && (
                <p className="text-sm text-red-500 mt-1">{formState.errors.status.message}</p>
              )}
            </Field>

            <Controller
              control={control}
              name="priority"
              render={({ field }) => (
                <div>
                  <Label>
                    {__("Priority")}
                    {" "}
                    *
                  </Label>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    {priorityOptions.map(option => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                  {formState.errors.priority?.message && (
                    <div className="text-red-500 text-sm mt-1">
                      {formState.errors.priority.message}
                    </div>
                  )}
                </div>
              )}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="identifiedOn">{__("Date Identified")}</Label>
              <Input
                id="identifiedOn"
                type="date"
                {...register("identifiedOn")}
              />
              {formState.errors.identifiedOn && (
                <p className="text-sm text-red-500">{formState.errors.identifiedOn.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="dueDate">{__("Due Date")}</Label>
              <Input
                id="dueDate"
                type="date"
                {...register("dueDate")}
              />
              {formState.errors.dueDate && (
                <p className="text-sm text-red-500">{formState.errors.dueDate.message}</p>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="rootCause">{__("Root Cause")}</Label>
            <Textarea
              id="rootCause"
              {...register("rootCause")}
              placeholder={__("Detailed analysis of the root cause...")}
              rows={3}
            />
            {formState.errors.rootCause && (
              <p className="text-sm text-red-500">{formState.errors.rootCause.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="correctiveAction">{__("Corrective Action")}</Label>
            <Textarea
              id="correctiveAction"
              {...register("correctiveAction")}
              placeholder={__("Proposed corrective actions...")}
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="effectivenessCheck">{__("Effectiveness Check")}</Label>
            <Textarea
              id="effectivenessCheck"
              {...register("effectivenessCheck")}
              placeholder={__("Assessment of corrective action effectiveness...")}
              rows={2}
            />
          </div>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={formState.isSubmitting}>
            {formState.isSubmitting ? __("Creating...") : __("Create Finding")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
