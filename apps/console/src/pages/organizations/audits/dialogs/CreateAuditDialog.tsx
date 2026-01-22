import {
  auditStates,
  formatDatetime,
  formatError,
  getAuditStateLabel,
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
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { Suspense } from "react";
import { type Control, Controller } from "react-hook-form";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateAuditDialogFrameworksQuery } from "#/__generated__/core/CreateAuditDialogFrameworksQuery.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { useCreateAudit } from "#/hooks/graph/AuditGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const frameworksQuery = graphql`
  query CreateAuditDialogFrameworksQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        frameworks(first: 100) {
          edges {
            node {
              id
              name
            }
          }
        }
      }
    }
  }
`;

const schema = z.object({
  frameworkId: z.string().min(1, "Framework is required"),
  name: z.string().optional(),
  validFrom: z.string().optional(),
  validUntil: z.string().optional(),
  state: z.enum([
    "NOT_STARTED",
    "IN_PROGRESS",
    "COMPLETED",
    "REJECTED",
    "OUTDATED",
  ]),
});

type Props = {
  children: React.ReactNode;
  connection: string;
  organizationId: string;
};

export function CreateAuditDialog({
  children,
  connection,
  organizationId,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(schema, {
      defaultValues: {
        frameworkId: "",
        name: "",
        validFrom: "",
        validUntil: "",
        state: "NOT_STARTED",
      },
    });
  const ref = useDialogRef();
  const createAudit = useCreateAudit(connection);

  const onSubmit = async (data: z.infer<typeof schema>) => {
    try {
      await createAudit({
        organizationId,
        frameworkId: data.frameworkId,
        name: data.name || null,
        validFrom: formatDatetime(data.validFrom),
        validUntil: formatDatetime(data.validUntil),
        state: data.state,
      });
      ref.current?.close();
      reset();
      toast({
        title: __("Success"),
        description: __("Audit created successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to create audit"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={<Breadcrumb items={[__("Audits"), __("New Audit")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          <Field label={__("Framework")}>
            <Suspense
              fallback={
                <Select variant="editor" disabled placeholder="Loading..." />
              }
            >
              <FrameworkSelect
                organizationId={organizationId}
                control={control}
                name="frameworkId"
              />
            </Suspense>
          </Field>

          <Field label={__("Name")}>
            <Input {...register("name")} placeholder={__("Audit name")} />
          </Field>

          <ControlledField
            control={control}
            name="state"
            type="select"
            label={__("State")}
          >
            {auditStates.map(state => (
              <Option key={state} value={state}>
                {getAuditStateLabel(__, state)}
              </Option>
            ))}
          </ControlledField>

          <Field label={__("Valid From")}>
            <Input {...register("validFrom")} type="date" />
          </Field>
          <Field label={__("Valid Until")}>
            <Input {...register("validUntil")} type="date" />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button disabled={formState.isSubmitting} type="submit">
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

type FormSchema = z.infer<typeof schema>;

function FrameworkSelect({
  organizationId,
  control,
  name,
}: {
  organizationId: string;
  control: Control<FormSchema>;
  name: keyof FormSchema;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<CreateAuditDialogFrameworksQuery>(
    frameworksQuery,
    { organizationId },
    { fetchPolicy: "network-only" },
  );
  const frameworks
    = data?.organization?.frameworks?.edges
      ?.map(edge => edge.node)
      .filter((node): node is NonNullable<typeof node> => node !== null) ?? [];

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Select
          id={name}
          variant="editor"
          placeholder={__("Select a framework")}
          onValueChange={field.onChange}
          {...field}
          className="w-full"
          value={field.value ?? ""}
        >
          {frameworks.map(framework => (
            <Option key={framework.id} value={framework.id}>
              {framework.name}
            </Option>
          ))}
        </Select>
      )}
    />
  );
}
