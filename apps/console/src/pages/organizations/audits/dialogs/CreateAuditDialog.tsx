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

import {
  auditStates,
  formatDatetime,
  formatError,
  type GraphQLError,
} from "@probo/helpers";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  type DialogRef,
  Field,
  IconUpload,
  Input,
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { Suspense } from "react";
import { type Control, Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
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

type Props = {
  children?: React.ReactNode;
  connection: string;
  organizationId: string;
  file?: File | null;
  ref?: DialogRef;
  onClose?: () => void;
};

export function CreateAuditDialog({
  children,
  connection,
  organizationId,
  file,
  ref: externalRef,
  onClose,
}: Props) {
  const { i18n, t } = useTranslation();
  const { toast } = useToast();
  const schema = z.object({
    frameworkId: z.string().min(1, t("createAuditDialog.validation.frameworkRequired")),
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
  const internalRef = useDialogRef();
  const ref = externalRef ?? internalRef;
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
        file: file ?? null,
      });

      ref.current?.close();
      reset();
      onClose?.();
      toast({
        title: t("createAuditDialog.messages.success"),
        description: file
          ? t("createAuditDialog.messages.createdWithReport")
          : t("createAuditDialog.messages.created"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: t("createAuditDialog.messages.error"),
        description: formatError(
          t("createAuditDialog.errors.create"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  };

  const handleClose = () => {
    onClose?.();
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={(
        <Breadcrumb items={[
          t("createAuditDialog.breadcrumb.audits"),
          t("createAuditDialog.breadcrumb.newAudit"),
        ]}
        />
      )}
      onClose={handleClose}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          {file && (
            <div className="flex items-center gap-3 rounded-lg border border-border-low bg-level-1 p-3">
              <IconUpload className="text-txt-secondary size-5 shrink-0" />
              <div className="min-w-0">
                <p className="text-txt-primary truncate text-sm font-medium">
                  {file.name}
                </p>
                <p className="text-txt-tertiary text-xs">
                  {t("createAuditDialog.fileSize", {
                    value: new Intl.NumberFormat(i18n.language, {
                      maximumFractionDigits: 2,
                    }).format(file.size / 1024 / 1024),
                  })}
                </p>
              </div>
            </div>
          )}

          <Field label={t("createAuditDialog.fields.framework")}>
            <Suspense
              fallback={(
                <Select
                  variant="editor"
                  disabled
                  placeholder={t("createAuditDialog.loading")}
                />
              )}
            >
              <FrameworkSelect
                organizationId={organizationId}
                control={control}
                name="frameworkId"
              />
            </Suspense>
          </Field>

          <Field label={t("createAuditDialog.fields.name")}>
            <Input
              {...register("name")}
              placeholder={t("createAuditDialog.fields.namePlaceholder")}
            />
          </Field>

          <ControlledField
            control={control}
            name="state"
            type="select"
            label={t("createAuditDialog.fields.state")}
          >
            {auditStates.map(state => (
              <Option key={state} value={state}>
                {t(`createAuditDialog.states.${state.toLowerCase()}`)}
              </Option>
            ))}
          </ControlledField>

          <Field label={t("createAuditDialog.fields.validFrom")}>
            <Input {...register("validFrom")} type="date" />
          </Field>
          <Field label={t("createAuditDialog.fields.validUntil")}>
            <Input {...register("validUntil")} type="date" />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button disabled={formState.isSubmitting} type="submit">
            {t("createAuditDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

type FormSchema = {
  frameworkId: string;
  name?: string;
  validFrom?: string;
  validUntil?: string;
  state: "NOT_STARTED" | "IN_PROGRESS" | "COMPLETED" | "REJECTED" | "OUTDATED";
};

function FrameworkSelect({
  organizationId,
  control,
  name,
}: {
  organizationId: string;
  control: Control<FormSchema>;
  name: keyof FormSchema;
}) {
  const { t } = useTranslation();
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
          placeholder={t("createAuditDialog.fields.frameworkPlaceholder")}
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
