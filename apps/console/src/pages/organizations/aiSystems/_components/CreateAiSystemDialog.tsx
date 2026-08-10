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

import { formatDatetime } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { CreateAiSystemDialogMutation } from "#/__generated__/core/CreateAiSystemDialogMutation.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { z } from "#/lib/zod";

import {
  AI_SYSTEM_COMPANY_ROLES,
  AI_SYSTEM_RISK_CLASSIFICATIONS,
  AI_SYSTEM_STATUSES,
  getCompanyRoleLabel,
  getRiskClassificationLabel,
  getStatusLabel,
} from "../_lib/aiSystemHelpers";

const createAiSystemMutation = graphql`
  mutation CreateAiSystemDialogMutation(
    $input: CreateAiSystemInput!
    $connections: [ID!]!
  ) {
    createAiSystem(input: $input) {
      aiSystemEdge @prependEdge(connections: $connections) {
        node {
          ...AiSystemListItem_aiSystem
        }
      }
    }
  }
`;

interface CreateAiSystemDialogProps {
  children: ReactNode;
  connectionIds?: string[];
}

export function CreateAiSystemDialog({
  children,
  connectionIds,
}: CreateAiSystemDialogProps) {
  const { t } = useTranslation();
  const prefix = "createAiSystemDialog";

  const schema = z.object({
    name: z.string().trim().min(1, t(`${prefix}.validation.nameRequired`)),
    version: z.string().optional(),
    companyRoles: z.array(z.enum(["PROVIDER", "DEPLOYER", "USER", "DEVELOPER"])),
    status: z.enum(["ACTIVE", "IN_DEVELOPMENT", "DECOMMISSIONED"]),
    ownerId: z.string().nullable().optional(),
    source: z.string().optional(),
    purpose: z.string().optional(),
    intendedUseCases: z.string().optional(),
    autonomyLevel: z.string().optional(),
    humanOversightMechanism: z.string().optional(),
    riskClassification: z.enum(["HIGH_RISK", "LIMITED", "MINIMAL", "GPAI"]),
    keyStakeholders: z.string().optional(),
    dataSourcesAndType: z.string().optional(),
    deploymentDate: z.string().optional(),
    lastReviewDate: z.string().optional(),
    nextReviewDate: z.string().optional(),
    notes: z.string().optional(),
  });

  type FormData = z.infer<typeof schema>;
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const { toast } = useToast();
  const [createAiSystem, isCreating] = useMutation<CreateAiSystemDialogMutation>(
    createAiSystemMutation,
    {
      errorToast: t(`${prefix}.errors.create`),
    },
  );

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      name: "",
      version: "",
      companyRoles: [],
      status: "ACTIVE" as const,
      ownerId: null,
      source: "",
      purpose: "",
      intendedUseCases: "",
      autonomyLevel: "",
      humanOversightMechanism: "",
      riskClassification: "MINIMAL" as const,
      keyStakeholders: "",
      dataSourcesAndType: "",
      deploymentDate: "",
      lastReviewDate: "",
      nextReviewDate: "",
      notes: "",
    },
  });

  const onSubmit = async (formData: FormData) => {
    await createAiSystem({
      variables: {
        input: {
          organizationId,
          name: formData.name,
          version: formData.version || undefined,
          companyRoles: formData.companyRoles,
          status: formData.status,
          ownerId: formData.ownerId || undefined,
          source: formData.source || undefined,
          purpose: formData.purpose || undefined,
          intendedUseCases: formData.intendedUseCases || undefined,
          autonomyLevel: formData.autonomyLevel || undefined,
          humanOversightMechanism: formData.humanOversightMechanism || undefined,
          riskClassification: formData.riskClassification,
          keyStakeholders: formData.keyStakeholders || undefined,
          dataSourcesAndType: formData.dataSourcesAndType || undefined,
          deploymentDate: formatDatetime(formData.deploymentDate),
          lastReviewDate: formatDatetime(formData.lastReviewDate),
          nextReviewDate: formatDatetime(formData.nextReviewDate),
          notes: formData.notes || undefined,
        },
        connections: connectionIds ?? [],
      },
    });
    toast({
      title: t(`${prefix}.messages.success`),
      description: t(`${prefix}.messages.created`),
      variant: "success",
    });
    reset();
    dialogRef.current?.close();
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={(
        <Breadcrumb
          items={[
            t(`${prefix}.breadcrumb.aiSystems`),
            t(`${prefix}.breadcrumb.create`),
          ]}
        />
      )}
      className="max-w-3xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4 max-h-[70vh] overflow-y-auto">
          <Field
            label={t(`${prefix}.fields.name`)}
            error={formState.errors.name?.message}
            required
          >
            <Input
              {...register("name")}
              placeholder={t(`${prefix}.fields.namePlaceholder`)}
            />
          </Field>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label={t(`${prefix}.fields.version`)}>
              <Input
                {...register("version")}
                placeholder={t(`${prefix}.fields.versionPlaceholder`)}
              />
            </Field>

            <Controller
              control={control}
              name="status"
              render={({ field }) => (
                <Field label={t(`${prefix}.fields.status`)} required>
                  <Select
                    variant="editor"
                    value={field.value}
                    onValueChange={field.onChange}
                    className="w-full"
                  >
                    {AI_SYSTEM_STATUSES.map(status => (
                      <Option key={status} value={status}>
                        {getStatusLabel(status, t, prefix)}
                      </Option>
                    ))}
                  </Select>
                </Field>
              )}
            />
          </div>

          <Controller
            control={control}
            name="companyRoles"
            render={({ field }) => (
              <Field label={t(`${prefix}.fields.companyRoles`)}>
                <div className="flex flex-wrap gap-4">
                  {AI_SYSTEM_COMPANY_ROLES.map(role => (
                    <label key={role} className="flex items-center gap-2 text-sm">
                      <input
                        type="checkbox"
                        checked={field.value.includes(role)}
                        onChange={() => {
                          const next = field.value.includes(role)
                            ? field.value.filter(value => value !== role)
                            : [...field.value, role];
                          field.onChange(next);
                        }}
                      />
                      {getCompanyRoleLabel(role, t, prefix)}
                    </label>
                  ))}
                </div>
              </Field>
            )}
          />

          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="ownerId"
            label={t(`${prefix}.fields.owner`)}
            error={formState.errors.ownerId?.message}
            optional
          />

          <Field label={t(`${prefix}.fields.source`)}>
            <Input
              {...register("source")}
              placeholder={t(`${prefix}.fields.sourcePlaceholder`)}
            />
          </Field>

          <Field label={t(`${prefix}.fields.purpose`)}>
            <Textarea
              {...register("purpose")}
              placeholder={t(`${prefix}.fields.purposePlaceholder`)}
              rows={2}
            />
          </Field>

          <Field label={t(`${prefix}.fields.intendedUseCases`)}>
            <Textarea
              {...register("intendedUseCases")}
              placeholder={t(`${prefix}.fields.intendedUseCasesPlaceholder`)}
              rows={2}
            />
          </Field>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label={t(`${prefix}.fields.autonomyLevel`)}>
              <Input
                {...register("autonomyLevel")}
                placeholder={t(`${prefix}.fields.autonomyLevelPlaceholder`)}
              />
            </Field>

            <Controller
              control={control}
              name="riskClassification"
              render={({ field }) => (
                <Field label={t(`${prefix}.fields.riskClassification`)} required>
                  <Select
                    variant="editor"
                    value={field.value}
                    onValueChange={field.onChange}
                    className="w-full"
                  >
                    {AI_SYSTEM_RISK_CLASSIFICATIONS.map(classification => (
                      <Option key={classification} value={classification}>
                        {getRiskClassificationLabel(classification, t, prefix)}
                      </Option>
                    ))}
                  </Select>
                </Field>
              )}
            />
          </div>

          <Field label={t(`${prefix}.fields.humanOversightMechanism`)}>
            <Textarea
              {...register("humanOversightMechanism")}
              placeholder={t(`${prefix}.fields.humanOversightMechanismPlaceholder`)}
              rows={2}
            />
          </Field>

          <Field label={t(`${prefix}.fields.keyStakeholders`)}>
            <Textarea
              {...register("keyStakeholders")}
              placeholder={t(`${prefix}.fields.keyStakeholdersPlaceholder`)}
              rows={2}
            />
          </Field>

          <Field label={t(`${prefix}.fields.dataSourcesAndType`)}>
            <Textarea
              {...register("dataSourcesAndType")}
              placeholder={t(`${prefix}.fields.dataSourcesAndTypePlaceholder`)}
              rows={2}
            />
          </Field>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Field label={t(`${prefix}.fields.deploymentDate`)}>
              <Input {...register("deploymentDate")} type="date" />
            </Field>
            <Field label={t(`${prefix}.fields.lastReviewDate`)}>
              <Input {...register("lastReviewDate")} type="date" />
            </Field>
            <Field label={t(`${prefix}.fields.nextReviewDate`)}>
              <Input {...register("nextReviewDate")} type="date" />
            </Field>
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">{t(`${prefix}.fields.notes`)}</Label>
            <Textarea
              id="notes"
              {...register("notes")}
              placeholder={t(`${prefix}.fields.notesPlaceholder`)}
              rows={2}
            />
          </div>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={formState.isSubmitting || isCreating}>
            {formState.isSubmitting || isCreating
              ? t(`${prefix}.actions.creating`)
              : t(`${prefix}.actions.create`)}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
