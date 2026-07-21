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
import { z } from "zod";

import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import { useCreateObligation } from "../../../../hooks/graph/ObligationGraph";

type FormData = {
  area?: string;
  source?: string;
  requirement?: string;
  actionsToBeImplemented?: string;
  regulator?: string;
  type: "LEGAL" | "CONTRACTUAL";
  ownerId: string;
  lastReviewDate?: string;
  dueDate?: string;
  status: "NON_COMPLIANT" | "PARTIALLY_COMPLIANT" | "COMPLIANT";
};

interface CreateObligationDialogProps {
  children: ReactNode;
  organizationId: string;
  connection?: string;
}

export function CreateObligationDialog({
  children,
  organizationId,
  connection,
}: CreateObligationDialogProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createObligation = useCreateObligation(connection || "");
  const statusOptions = ["NON_COMPLIANT", "PARTIALLY_COMPLIANT", "COMPLIANT"] as const;
  const typeOptions = ["LEGAL", "CONTRACTUAL"] as const;
  const schema = z.object({
    area: z.string().optional(),
    source: z.string().optional(),
    requirement: z.string().optional(),
    actionsToBeImplemented: z.string().optional(),
    regulator: z.string().optional(),
    type: z.enum(typeOptions),
    ownerId: z.string().min(1, t("createObligationDialog.validation.ownerRequired")),
    lastReviewDate: z.string().optional(),
    dueDate: z.string().optional(),
    status: z.enum(statusOptions),
  });

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      area: "",
      source: "",
      requirement: "",
      actionsToBeImplemented: "",
      regulator: "",
      type: "LEGAL" as const,
      ownerId: "",
      lastReviewDate: "",
      dueDate: "",
      status: "NON_COMPLIANT" as const,
    },
  });

  const onSubmit = async (formData: FormData) => {
    try {
      await createObligation({
        organizationId,
        area: formData.area || undefined,
        source: formData.source || undefined,
        requirement: formData.requirement || undefined,
        actionsToBeImplemented: formData.actionsToBeImplemented || undefined,
        regulator: formData.regulator || undefined,
        type: formData.type,
        ownerId: formData.ownerId,
        lastReviewDate: formatDatetime(formData.lastReviewDate),
        dueDate: formatDatetime(formData.dueDate),
        status: formData.status,
      });

      toast({
        title: t("createObligationDialog.messages.success"),
        description: t("createObligationDialog.messages.created"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: t("createObligationDialog.messages.error"),
        description: formatError(t("createObligationDialog.errors.create"), error as GraphQLError),
        variant: "error",
      });
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[t("createObligationDialog.breadcrumb.obligations"), t("createObligationDialog.breadcrumb.create")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Field
              label={t("createObligationDialog.fields.area")}
              {...register("area")}
              placeholder={t("createObligationDialog.placeholders.area")}
              error={formState.errors.area?.message}
            />

            <Field
              label={t("createObligationDialog.fields.source")}
              {...register("source")}
              placeholder={t("createObligationDialog.placeholders.source")}
              error={formState.errors.source?.message}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Field label={t("createObligationDialog.fields.status")}>
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <Select
                    variant="editor"
                    placeholder={t("createObligationDialog.placeholders.status")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                  >
                    {statusOptions.map(option => (
                      <Option key={option} value={option}>
                        {t(`createObligationDialog.statuses.${option.toLowerCase()}`)}
                      </Option>
                    ))}
                  </Select>
                )}
              />
              {formState.errors.status && (
                <p className="text-sm text-red-500 mt-1">{formState.errors.status.message}</p>
              )}
            </Field>

            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={t("createObligationDialog.fields.owner")}
              error={formState.errors.ownerId?.message}
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Field
              label={t("createObligationDialog.fields.regulator")}
              {...register("regulator")}
              placeholder={t("createObligationDialog.placeholders.regulator")}
              error={formState.errors.regulator?.message}
            />

            <Field label={t("createObligationDialog.fields.type")}>
              <Controller
                control={control}
                name="type"
                render={({ field }) => (
                  <Select
                    variant="editor"
                    placeholder={t("createObligationDialog.placeholders.type")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                  >
                    {typeOptions.map(option => (
                      <Option key={option} value={option}>
                        {t(`createObligationDialog.types.${option.toLowerCase()}`)}
                      </Option>
                    ))}
                  </Select>
                )}
              />
              {formState.errors.type && (
                <p className="text-sm text-red-500 mt-1">{formState.errors.type.message}</p>
              )}
            </Field>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="lastReviewDate">{t("createObligationDialog.fields.lastReviewDate")}</Label>
              <Input
                id="lastReviewDate"
                type="date"
                {...register("lastReviewDate")}
              />
              {formState.errors.lastReviewDate && (
                <p className="text-sm text-red-500">{formState.errors.lastReviewDate.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="dueDate">{t("createObligationDialog.fields.dueDate")}</Label>
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
            <Label htmlFor="requirement">{t("createObligationDialog.fields.requirement")}</Label>
            <Textarea
              id="requirement"
              {...register("requirement")}
              placeholder={t("createObligationDialog.placeholders.requirement")}
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="actionsToBeImplemented">{t("createObligationDialog.fields.actionsToBeImplemented")}</Label>
            <Textarea
              id="actionsToBeImplemented"
              {...register("actionsToBeImplemented")}
              placeholder={t("createObligationDialog.placeholders.actionsToBeImplemented")}
              rows={3}
            />
          </div>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={formState.isSubmitting}>
            {formState.isSubmitting ? t("createObligationDialog.actions.creating") : t("createObligationDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
