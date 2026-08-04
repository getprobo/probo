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
  Field,
  Input,
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, Suspense } from "react";
import { type Control, Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateAuditProgramDialogFrameworksQuery } from "#/__generated__/core/CreateAuditProgramDialogFrameworksQuery.graphql";
import { useCreateAuditProgram } from "#/hooks/graph/AuditProgramGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const frameworksQuery = graphql`
  query CreateAuditProgramDialogFrameworksQuery($organizationId: ID!) {
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
  children?: ReactNode;
  connection: string;
  organizationId: string;
};

export function CreateAuditProgramDialog({
  children,
  connection,
  organizationId,
}: Props) {
  const { t } = useTranslation("organizations/audit-programs");
  const { toast } = useToast();
  const schema = z.object({
    frameworkId: z
      .string()
      .min(1, t("createAuditProgramDialog.validation.frameworkRequired")),
    name: z
      .string()
      .trim()
      .min(1, t("createAuditProgramDialog.validation.nameRequired")),
    validFrom: z.string().optional(),
    validUntil: z.string().optional(),
  });
  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(schema, {
      defaultValues: {
        frameworkId: "",
        name: "",
        validFrom: "",
        validUntil: "",
      },
    });
  const ref = useDialogRef();
  const createAuditProgram = useCreateAuditProgram(connection);

  const onSubmit = async (data: z.infer<typeof schema>) => {
    try {
      await createAuditProgram({
        organizationId,
        frameworkId: data.frameworkId,
        name: data.name,
        validFrom: formatDatetime(data.validFrom),
        validUntil: formatDatetime(data.validUntil),
      });

      ref.current?.close();
      reset();
      toast({
        title: t("createAuditProgramDialog.messages.success"),
        description: t("createAuditProgramDialog.messages.created"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: t("createAuditProgramDialog.messages.error"),
        description: formatError(
          t("createAuditProgramDialog.errors.create"),
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
      title={(
        <Breadcrumb
          items={[
            t("createAuditProgramDialog.breadcrumb.auditPrograms"),
            t("createAuditProgramDialog.breadcrumb.new"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          <Field label={t("createAuditProgramDialog.fields.framework")}>
            <Suspense
              fallback={(
                <Select
                  variant="editor"
                  disabled
                  placeholder={t("createAuditProgramDialog.loading")}
                />
              )}
            >
              <FrameworkSelect organizationId={organizationId} control={control} />
            </Suspense>
          </Field>

          <Field label={t("createAuditProgramDialog.fields.name")}>
            <Input
              {...register("name")}
              placeholder={t("createAuditProgramDialog.fields.namePlaceholder")}
            />
          </Field>

          <Field label={t("createAuditProgramDialog.fields.validFrom")}>
            <Input {...register("validFrom")} type="date" />
          </Field>
          <Field label={t("createAuditProgramDialog.fields.validUntil")}>
            <Input {...register("validUntil")} type="date" />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button disabled={formState.isSubmitting} type="submit">
            {t("createAuditProgramDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

type FormSchema = {
  frameworkId: string;
  name: string;
  validFrom?: string;
  validUntil?: string;
};

function FrameworkSelect({
  organizationId,
  control,
}: {
  organizationId: string;
  control: Control<FormSchema>;
}) {
  const { t } = useTranslation("organizations/audit-programs");
  const data = useLazyLoadQuery<CreateAuditProgramDialogFrameworksQuery>(
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
      name="frameworkId"
      render={({ field }) => (
        <Select
          id="frameworkId"
          variant="editor"
          placeholder={t("createAuditProgramDialog.fields.frameworkPlaceholder")}
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
