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
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  Field,
  IconTrashCan,
  Input,
  Option,
  Select,
  Textarea,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";

import type { AiSystemDetailsPageDeleteMutation } from "#/__generated__/core/AiSystemDetailsPageDeleteMutation.graphql";
import type { AiSystemDetailsPageQuery } from "#/__generated__/core/AiSystemDetailsPageQuery.graphql";
import type { AiSystemDetailsPageUpdateMutation } from "#/__generated__/core/AiSystemDetailsPageUpdateMutation.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";
import { z } from "#/lib/zod";

import {
  AI_SYSTEM_COMPANY_ROLES,
  AI_SYSTEM_RISK_CLASSIFICATIONS,
  AI_SYSTEM_STATUSES,
  aiSystemListConnectionFilters,
  AiSystemsConnectionKey,
  getCompanyRoleLabel,
  getRiskClassificationLabel,
  getRiskClassificationVariant,
  getStatusLabel,
  getStatusVariant,
} from "./_lib/aiSystemHelpers";

export const aiSystemDetailsPageQuery = graphql`
  query AiSystemDetailsPageQuery($aiSystemId: ID!) {
    node(id: $aiSystemId) @required(action: THROW) {
      __typename
      ... on AiSystem {
        id
        name
        version
        companyRoles
        status
        source
        purpose
        intendedUseCases
        autonomyLevel
        humanOversightMechanism
        riskClassification
        keyStakeholders
        dataSourcesAndType
        deploymentDate
        lastReviewDate
        nextReviewDate
        notes
        owner {
          id
        }
        canUpdate: permission(action: "core:ai-system:update")
        canDelete: permission(action: "core:ai-system:delete")
      }
    }
  }
`;

const updateAiSystemMutation = graphql`
  mutation AiSystemDetailsPageUpdateMutation($input: UpdateAiSystemInput!) {
    updateAiSystem(input: $input) {
      aiSystem {
        id
        name
        version
        companyRoles
        status
        source
        purpose
        intendedUseCases
        autonomyLevel
        humanOversightMechanism
        riskClassification
        keyStakeholders
        dataSourcesAndType
        deploymentDate
        lastReviewDate
        nextReviewDate
        notes
        owner {
          id
          fullName
        }
        updatedAt
      }
    }
  }
`;

const deleteAiSystemMutation = graphql`
  mutation AiSystemDetailsPageDeleteMutation(
    $input: DeleteAiSystemInput!
    $connections: [ID!]!
  ) {
    deleteAiSystem(input: $input) {
      deletedAiSystemId @deleteEdge(connections: $connections)
    }
  }
`;

interface AiSystemDetailsPageProps {
  queryRef: PreloadedQuery<AiSystemDetailsPageQuery>;
}

export function AiSystemDetailsPage({ queryRef }: AiSystemDetailsPageProps) {
  const { node: aiSystem } = usePreloadedQuery<AiSystemDetailsPageQuery>(
    aiSystemDetailsPageQuery,
    queryRef,
  );
  if (aiSystem.__typename !== "AiSystem") {
    throw new NotFoundError("AI system not found");
  }

  const { t } = useTranslation();
  const prefix = "aiSystemDetailsPage";
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const { toast } = useToast();

  const updateAiSystemSchema = z.object({
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

  const [updateAiSystem] = useMutation<AiSystemDetailsPageUpdateMutation>(
    updateAiSystemMutation,
    {
      errorToast: t(`${prefix}.errors.update`),
    },
  );
  const [deleteAiSystem] = useMutation<AiSystemDetailsPageDeleteMutation>(
    deleteAiSystemMutation,
    {
      errorToast: t(`${prefix}.errors.delete`),
    },
  );

  const connections = aiSystemListConnectionFilters(aiSystem).map(filter =>
    ConnectionHandler.getConnectionID(
      organizationId,
      AiSystemsConnectionKey,
      { filter },
    ),
  );

  const handleDelete = () => {
    confirm(
      async () => {
        await deleteAiSystem({
          variables: {
            input: { aiSystemId: aiSystem.id },
            connections,
          },
        });
        toast({
          title: t(`${prefix}.messages.success`),
          description: t(`${prefix}.messages.deleted`),
          variant: "success",
        });
        void navigate(`/organizations/${organizationId}/settings/ai-systems`);
      },
      {
        message: t(`${prefix}.deleteConfirmation`, {
          name: aiSystem.name,
        }),
      },
    );
  };

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateAiSystemSchema, {
      defaultValues: {
        name: aiSystem.name || "",
        version: aiSystem.version || "",
        companyRoles: [...(aiSystem.companyRoles ?? [])],
        status: aiSystem.status || "ACTIVE",
        ownerId: aiSystem.owner?.id ?? null,
        source: aiSystem.source || "",
        purpose: aiSystem.purpose || "",
        intendedUseCases: aiSystem.intendedUseCases || "",
        autonomyLevel: aiSystem.autonomyLevel || "",
        humanOversightMechanism: aiSystem.humanOversightMechanism || "",
        riskClassification: aiSystem.riskClassification,
        keyStakeholders: aiSystem.keyStakeholders || "",
        dataSourcesAndType: aiSystem.dataSourcesAndType || "",
        deploymentDate: aiSystem.deploymentDate?.split("T")[0] || "",
        lastReviewDate: aiSystem.lastReviewDate?.split("T")[0] || "",
        nextReviewDate: aiSystem.nextReviewDate?.split("T")[0] || "",
        notes: aiSystem.notes || "",
      },
    });

  const onSubmit = handleSubmit(async (formData) => {
    const { dirtyFields } = formState;

    await updateAiSystem({
      variables: {
        input: {
          id: aiSystem.id,
          ...(dirtyFields.name ? { name: formData.name } : {}),
          ...(dirtyFields.version ? { version: formData.version || null } : {}),
          ...(dirtyFields.companyRoles ? { companyRoles: formData.companyRoles } : {}),
          ...(dirtyFields.status ? { status: formData.status } : {}),
          ...(dirtyFields.ownerId ? { ownerId: formData.ownerId || null } : {}),
          ...(dirtyFields.source ? { source: formData.source || null } : {}),
          ...(dirtyFields.purpose ? { purpose: formData.purpose || null } : {}),
          ...(dirtyFields.intendedUseCases
            ? { intendedUseCases: formData.intendedUseCases || null }
            : {}),
          ...(dirtyFields.autonomyLevel
            ? { autonomyLevel: formData.autonomyLevel || null }
            : {}),
          ...(dirtyFields.humanOversightMechanism
            ? { humanOversightMechanism: formData.humanOversightMechanism || null }
            : {}),
          ...(dirtyFields.riskClassification
            ? { riskClassification: formData.riskClassification }
            : {}),
          ...(dirtyFields.keyStakeholders
            ? { keyStakeholders: formData.keyStakeholders || null }
            : {}),
          ...(dirtyFields.dataSourcesAndType
            ? { dataSourcesAndType: formData.dataSourcesAndType || null }
            : {}),
          ...(dirtyFields.deploymentDate
            ? { deploymentDate: formatDatetime(formData.deploymentDate) ?? null }
            : {}),
          ...(dirtyFields.lastReviewDate
            ? { lastReviewDate: formatDatetime(formData.lastReviewDate) ?? null }
            : {}),
          ...(dirtyFields.nextReviewDate
            ? { nextReviewDate: formatDatetime(formData.nextReviewDate) ?? null }
            : {}),
          ...(dirtyFields.notes ? { notes: formData.notes || null } : {}),
        },
      },
    });
    reset(formData);
    toast({
      title: t(`${prefix}.messages.success`),
      description: t(`${prefix}.messages.updated`),
      variant: "success",
    });
  });

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t(`${prefix}.breadcrumb.aiSystems`),
            to: `/organizations/${organizationId}/settings/ai-systems`,
          },
          {
            label: aiSystem.name || t(`${prefix}.breadcrumb.unknown`),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4 flex-wrap">
          <div className="text-2xl font-semibold">{aiSystem.name}</div>
          <Badge variant={getStatusVariant(aiSystem.status)}>
            {getStatusLabel(aiSystem.status, t, prefix)}
          </Badge>
          {aiSystem.riskClassification && (
            <Badge variant={getRiskClassificationVariant(aiSystem.riskClassification)}>
              {getRiskClassificationLabel(
                aiSystem.riskClassification,
                t,
                prefix,
              )}
            </Badge>
          )}
        </div>
        {aiSystem.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {t(`${prefix}.actions.delete`)}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <div className="max-w-4xl">
        <Card padded>
          <form onSubmit={e => void onSubmit(e)} className="space-y-6">
            <Field
              label={t(`${prefix}.fields.name`)}
              error={formState.errors.name?.message}
              required
            >
              <Input
                {...register("name")}
                placeholder={t(`${prefix}.fields.namePlaceholder`)}
                disabled={!aiSystem.canUpdate}
              />
            </Field>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field label={t(`${prefix}.fields.version`)}>
                <Input
                  {...register("version")}
                  placeholder={t(`${prefix}.fields.versionPlaceholder`)}
                  disabled={!aiSystem.canUpdate}
                />
              </Field>

              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <Field label={t(`${prefix}.fields.status`)} required>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      disabled={!aiSystem.canUpdate}
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
                          disabled={!aiSystem.canUpdate}
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
              disabled={!aiSystem.canUpdate}
            />

            <Field label={t(`${prefix}.fields.source`)}>
              <Input
                {...register("source")}
                placeholder={t(`${prefix}.fields.sourcePlaceholder`)}
                disabled={!aiSystem.canUpdate}
              />
            </Field>

            <Field label={t(`${prefix}.fields.purpose`)}>
              <Textarea
                {...register("purpose")}
                placeholder={t(`${prefix}.fields.purposePlaceholder`)}
                rows={3}
                disabled={!aiSystem.canUpdate}
              />
            </Field>

            <Field label={t(`${prefix}.fields.intendedUseCases`)}>
              <Textarea
                {...register("intendedUseCases")}
                placeholder={t(`${prefix}.fields.intendedUseCasesPlaceholder`)}
                rows={3}
                disabled={!aiSystem.canUpdate}
              />
            </Field>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field label={t(`${prefix}.fields.autonomyLevel`)}>
                <Input
                  {...register("autonomyLevel")}
                  placeholder={t(`${prefix}.fields.autonomyLevelPlaceholder`)}
                  disabled={!aiSystem.canUpdate}
                />
              </Field>

              <Controller
                control={control}
                name="riskClassification"
                render={({ field }) => (
                  <Field label={t(`${prefix}.fields.riskClassification`)} required>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      disabled={!aiSystem.canUpdate}
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
                rows={3}
                disabled={!aiSystem.canUpdate}
              />
            </Field>

            <Field label={t(`${prefix}.fields.keyStakeholders`)}>
              <Textarea
                {...register("keyStakeholders")}
                placeholder={t(`${prefix}.fields.keyStakeholdersPlaceholder`)}
                rows={3}
                disabled={!aiSystem.canUpdate}
              />
            </Field>

            <Field label={t(`${prefix}.fields.dataSourcesAndType`)}>
              <Textarea
                {...register("dataSourcesAndType")}
                placeholder={t(`${prefix}.fields.dataSourcesAndTypePlaceholder`)}
                rows={3}
                disabled={!aiSystem.canUpdate}
              />
            </Field>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Field label={t(`${prefix}.fields.deploymentDate`)}>
                <Input
                  {...register("deploymentDate")}
                  type="date"
                  disabled={!aiSystem.canUpdate}
                />
              </Field>
              <Field label={t(`${prefix}.fields.lastReviewDate`)}>
                <Input
                  {...register("lastReviewDate")}
                  type="date"
                  disabled={!aiSystem.canUpdate}
                />
              </Field>
              <Field label={t(`${prefix}.fields.nextReviewDate`)}>
                <Input
                  {...register("nextReviewDate")}
                  type="date"
                  disabled={!aiSystem.canUpdate}
                />
              </Field>
            </div>

            <Field label={t(`${prefix}.fields.notes`)}>
              <Textarea
                {...register("notes")}
                placeholder={t(`${prefix}.fields.notesPlaceholder`)}
                rows={3}
                disabled={!aiSystem.canUpdate}
              />
            </Field>

            <div className="flex justify-end">
              {formState.isDirty && aiSystem.canUpdate && (
                <Button type="submit" disabled={formState.isSubmitting}>
                  {formState.isSubmitting
                    ? t(`${prefix}.actions.updating`)
                    : t(`${prefix}.actions.update`)}
                </Button>
              )}
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
}
