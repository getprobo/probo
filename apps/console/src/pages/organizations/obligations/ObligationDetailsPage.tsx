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

import { formatError, type GraphQLError } from "@probo/helpers";
import {
  formatDatetime,
  getObligationStatusVariant,
} from "@probo/helpers";
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
  useToast,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { z } from "zod";

import type { ObligationGraphNodeQuery } from "#/__generated__/core/ObligationGraphNodeQuery.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  obligationNodeQuery,
  ObligationsConnectionKey,
  useDeleteObligation,
  useUpdateObligation,
} from "../../../hooks/graph/ObligationGraph";

type Props = {
  queryRef: PreloadedQuery<ObligationGraphNodeQuery>;
};

export default function ObligationDetailsPage(props: Props) {
  const { queryRef } = props;
  const { node: obligation } = usePreloadedQuery<ObligationGraphNodeQuery>(
    obligationNodeQuery,
    queryRef,
  );
  const { t } = useTranslation();
  const { toast } = useToast();
  const organizationId = useOrganizationId();

  const disabled = !obligation.canUpdate;

  const updateObligation = useUpdateObligation();
  const statusOptions = ["NON_COMPLIANT", "PARTIALLY_COMPLIANT", "COMPLIANT"] as const;
  const typeOptions = ["LEGAL", "CONTRACTUAL"] as const;
  const updateObligationSchema = z.object({
    area: z.string().optional(),
    source: z.string().optional(),
    requirement: z.string().optional(),
    actionsToBeImplemented: z.string().optional(),
    regulator: z.string().optional(),
    type: z.enum(typeOptions),
    lastReviewDate: z.string().optional(),
    dueDate: z.string().optional(),
    status: z.enum(statusOptions),
    ownerId: z.string().min(1, t("obligationDetailsPage.validation.ownerRequired")),
  });

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ObligationsConnectionKey,
  );

  const deleteObligation = useDeleteObligation(
    { id: obligation?.id ?? "" },
    connectionId,
  );

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateObligationSchema,
    {
      defaultValues: {
        area: obligation?.area || "",
        source: obligation?.source || "",
        requirement: obligation?.requirement || "",
        actionsToBeImplemented:
                    obligation?.actionsToBeImplemented || "",
        regulator: obligation?.regulator || "",
        type: obligation?.type ?? "LEGAL",
        lastReviewDate: obligation?.lastReviewDate
          ? new Date(obligation.lastReviewDate)
            .toISOString()
            .split("T")[0]
          : "",
        dueDate: obligation?.dueDate
          ? new Date(obligation.dueDate).toISOString().split("T")[0]
          : "",
        status: obligation?.status ?? "NON_COMPLIANT",
        ownerId: obligation?.owner?.id || "",
      },
    },
  );

  const onSubmit = handleSubmit(async (formData) => {
    try {
      await updateObligation({
        id: obligation.id!,
        area: formData.area || undefined,
        source: formData.source || undefined,
        requirement: formData.requirement || undefined,
        actionsToBeImplemented:
                    formData.actionsToBeImplemented || undefined,
        regulator: formData.regulator || undefined,
        type: formData.type,
        lastReviewDate: formatDatetime(formData.lastReviewDate) ?? null,
        dueDate: formatDatetime(formData.dueDate) ?? null,
        status: formData.status,
        ownerId: formData.ownerId,
      });

      toast({
        title: t("obligationDetailsPage.messages.success"),
        description: t("obligationDetailsPage.messages.updated"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: t("obligationDetailsPage.messages.error"),
        description: formatError(
          t("obligationDetailsPage.errors.update"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const breadcrumbObligationsUrl = `/organizations/${organizationId}/obligations`;

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-start">
        <div>
          <Breadcrumb
            items={[
              {
                label: t("obligationDetailsPage.breadcrumb.obligations"),
                to: breadcrumbObligationsUrl,
              },
              { label: t("obligationDetailsPage.breadcrumb.details") },
            ]}
          />
          <div className="flex items-center gap-3 mt-2">
            <h1 className="text-2xl font-bold">
              {t("obligationDetailsPage.title")}
            </h1>
            <Badge
              variant={getObligationStatusVariant(
                obligation.status ?? "NON_COMPLIANT",
              )}
            >
              {t(`obligationDetailsPage.statuses.${(obligation.status ?? "NON_COMPLIANT").toLowerCase()}`)}
            </Badge>
          </div>
        </div>

        {obligation.canDelete && (
          <ActionDropdown>
            <DropdownItem
              icon={IconTrashCan}
              onClick={deleteObligation}
            >
              {t("obligationDetailsPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <Card padded>
        <form onSubmit={e => void onSubmit(e)} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field
              label={t("obligationDetailsPage.fields.area")}
              error={formState.errors.area?.message}
            >
              <Input
                {...register("area")}
                placeholder={t("obligationDetailsPage.placeholders.area")}
                disabled={disabled}
              />
            </Field>

            <Field
              label={t("obligationDetailsPage.fields.source")}
              error={formState.errors.source?.message}
            >
              <Input
                {...register("source")}
                placeholder={t("obligationDetailsPage.placeholders.source")}
                disabled={disabled}
              />
            </Field>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label={t("obligationDetailsPage.fields.status")}>
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <Select
                    variant="editor"
                    placeholder={t("obligationDetailsPage.placeholders.status")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                    disabled={disabled}
                  >
                    {statusOptions.map(option => (
                      <Option
                        key={option}
                        value={option}
                      >
                        {t(`obligationDetailsPage.statuses.${option.toLowerCase()}`)}
                      </Option>
                    ))}
                  </Select>
                )}
              />
              {formState.errors.status && (
                <p className="text-sm text-red-500 mt-1">
                  {formState.errors.status.message}
                </p>
              )}
            </Field>

            <Controller
              name="ownerId"
              control={control}
              render={() => (
                <PeopleSelectField
                  organizationId={organizationId}
                  control={control}
                  name="ownerId"
                  label={t("obligationDetailsPage.fields.owner")}
                  error={formState.errors.ownerId?.message}
                  required
                  disabled={disabled}
                />
              )}
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field
              label={t("obligationDetailsPage.fields.regulator")}
              error={formState.errors.regulator?.message}
            >
              <Input
                {...register("regulator")}
                placeholder={t("obligationDetailsPage.placeholders.regulator")}
                disabled={disabled}
              />
            </Field>

            <Field
              label={t("obligationDetailsPage.fields.type")}
              error={formState.errors.type?.message}
            >
              <Controller
                control={control}
                name="type"
                render={({ field }) => (
                  <Select
                    variant="editor"
                    placeholder={t("obligationDetailsPage.placeholders.type")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                    disabled={disabled}
                  >
                    {typeOptions.map(option => (
                      <Option
                        key={option}
                        value={option}
                      >
                        {t(`obligationDetailsPage.types.${option.toLowerCase()}`)}
                      </Option>
                    ))}
                  </Select>
                )}
              />
            </Field>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field
              label={t("obligationDetailsPage.fields.lastReviewDate")}
              error={formState.errors.lastReviewDate?.message}
            >
              <Input
                {...register("lastReviewDate")}
                type="date"
                disabled={disabled}
              />
            </Field>

            <Field
              label={t("obligationDetailsPage.fields.dueDate")}
              error={formState.errors.dueDate?.message}
            >
              <Input
                {...register("dueDate")}
                type="date"
                disabled={disabled}
              />
            </Field>
          </div>

          <Field
            label={t("obligationDetailsPage.fields.requirement")}
            error={formState.errors.requirement?.message}
          >
            <Textarea
              {...register("requirement")}
              placeholder={t("obligationDetailsPage.placeholders.requirement")}
              rows={4}
              disabled={disabled}
            />
          </Field>

          <Field
            label={t("obligationDetailsPage.fields.actionsToBeImplemented")}
            error={formState.errors.actionsToBeImplemented?.message}
          >
            <Textarea
              {...register("actionsToBeImplemented")}
              placeholder={t("obligationDetailsPage.placeholders.actionsToBeImplemented")}
              rows={4}
              disabled={disabled}
            />
          </Field>

          <div className="flex justify-end">
            {obligation.canUpdate && (
              <Button
                type="submit"
                disabled={formState.isSubmitting}
              >
                {formState.isSubmitting
                  ? t("obligationDetailsPage.actions.saving")
                  : t("obligationDetailsPage.actions.save")}
              </Button>
            )}
          </div>
        </form>
      </Card>
    </div>
  );
}
