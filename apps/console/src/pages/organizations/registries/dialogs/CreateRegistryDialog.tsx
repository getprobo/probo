import { type ReactNode, Suspense } from "react";
import {
  Button,
  Field,
  useToast,
  Dialog,
  DialogContent,
  DialogFooter,
  useDialogRef,
  Textarea,
  Breadcrumb,
  Label,
  Select,
  Option,
  Input,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useCreateNonconformityRegistry } from "../../../../hooks/graph/NonconformityRegistryGraph";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { Controller, type Control } from "react-hook-form";
import { useLazyLoadQuery, graphql } from "react-relay";
import { formatDatetime } from "@probo/helpers";

const auditsQuery = graphql`
  query CreateRegistryDialogAuditsQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        audits(first: 100) {
          edges {
            node {
              id
              framework {
                id
                name
              }
              state
              validFrom
              validUntil
            }
          }
        }
      }
    }
  }
`;

const schema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  description: z.string().optional(),
  auditId: z.string().min(1, "Audit is required"),
  dateIdentified: z.string().optional(),
  rootCause: z.string().min(1, "Root cause is required"),
  correctiveAction: z.string().optional(),
  ownerId: z.string().min(1, "Owner is required"),
  dueDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  effectivenessCheck: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

interface CreateRegistryDialogProps {
  children: ReactNode;
  connection?: string;
  organizationId: string;
}

export function CreateRegistryDialog({
  children,
  organizationId,
  connection,
}: CreateRegistryDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createRegistry = useCreateNonconformityRegistry(connection || "");

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      referenceId: "",
      description: "",
      auditId: "",
      dateIdentified: "",
      rootCause: "",
      correctiveAction: "",
      ownerId: "",
      dueDate: "",
      status: "OPEN" as const,
      effectivenessCheck: "",
    },
  });

  const onSubmit = handleSubmit(async (formData: FormData) => {
    try {
      await createRegistry({
        organizationId,
        referenceId: formData.referenceId,
        description: formData.description || undefined,
        auditId: formData.auditId,
        dateIdentified: formatDatetime(formData.dateIdentified),
        rootCause: formData.rootCause,
        correctiveAction: formData.correctiveAction || undefined,
        ownerId: formData.ownerId,
        dueDate: formatDatetime(formData.dueDate),
        status: formData.status,
        effectivenessCheck: formData.effectivenessCheck || undefined,
      });

      toast({
        title: __("Success"),
        description: __("Registry entry created successfully"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: __("Error"),
        description: __("Failed to create registry entry"),
        variant: "error",
      });
    }
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Registries"), __("Create Entry")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Reference ID")}
            {...register("referenceId")}
            placeholder="NC-001"
            error={formState.errors.referenceId?.message}
            required
          />

          <Field label={__("Audit")}>
            <Suspense fallback={<Select variant="editor" disabled placeholder="Loading..." />}>
              <AuditSelect
                organizationId={organizationId}
                control={control}
                name="auditId"
              />
            </Suspense>
            {formState.errors.auditId && (
              <p className="text-sm text-red-500 mt-1">{formState.errors.auditId.message}</p>
            )}
          </Field>

          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="ownerId"
            label={__("Owner")}
            error={formState.errors.ownerId?.message}
            required
          />

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="dateIdentified">{__("Date Identified")}</Label>
              <Input
                id="dateIdentified"
                type="date"
                {...register("dateIdentified")}
              />
              {formState.errors.dateIdentified && (
                <p className="text-sm text-red-500">{formState.errors.dateIdentified.message}</p>
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
                  <Option value="OPEN">{__("Open")}</Option>
                  <Option value="IN_PROGRESS">{__("In Progress")}</Option>
                  <Option value="CLOSED">{__("Closed")}</Option>
                </Select>
              )}
            />
            {formState.errors.status && (
              <p className="text-sm text-red-500 mt-1">{formState.errors.status.message}</p>
            )}
          </Field>

          <div className="space-y-2">
            <Label htmlFor="description">{__("Description")}</Label>
            <Textarea
              id="description"
              {...register("description")}
              placeholder={__("Brief description of the nonconformity...")}
              rows={2}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="rootCause">{__("Root Cause")} *</Label>
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
            {formState.isSubmitting ? __("Creating...") : __("Create Registry Entry")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function AuditSelect({
  organizationId,
  control,
  name
}: {
  organizationId: string;
  control: Control<FormData>;
  name: keyof FormData;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery(auditsQuery, { organizationId }) as any;
  const audits = data?.organization?.audits?.edges?.map((edge: any) => edge.node).filter((node: any): node is NonNullable<typeof node> => node !== null) ?? [];

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Select
          id={name}
          variant="editor"
          placeholder={__("Select an audit")}
          onValueChange={field.onChange}
          {...field}
          className="w-full"
          value={field.value ?? ""}
        >
          {audits.map((audit: any) => (
            <Option key={audit.id} value={audit.id}>
              {audit.framework.name} ({audit.state})
            </Option>
          ))}
        </Select>
      )}
    />
  );
}
