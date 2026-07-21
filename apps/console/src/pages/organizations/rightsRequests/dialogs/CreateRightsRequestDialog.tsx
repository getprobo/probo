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

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import { useCreateRightsRequest } from "../../../../hooks/graph/RightsRequestGraph";

type FormData = { requestType: "ACCESS" | "DELETION" | "RECTIFICATION" | "PORTABILITY" | "OBJECTION" | "COMPLAINT"; requestState: "TODO" | "IN_PROGRESS" | "DONE" | "REJECTED"; dataSubject?: string; contact?: string; details?: string; deadline?: string; actionTaken?: string };

interface CreateRightsRequestDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionId?: string;
}

export function CreateRightsRequestDialog({
  children,
  organizationId,
  connectionId,
}: CreateRightsRequestDialogProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createRequest = useCreateRightsRequest(connectionId || "");
  const schema = z.object({
    requestType: z.enum(["ACCESS", "DELETION", "RECTIFICATION", "PORTABILITY", "OBJECTION", "COMPLAINT"]), requestState: z.enum(["TODO", "IN_PROGRESS", "DONE", "REJECTED"]), dataSubject: z.string().optional(), contact: z.string().optional(), details: z.string().optional(), deadline: z.string().optional(), actionTaken: z.string().optional(),
  });

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      requestType: "ACCESS" as const,
      requestState: "TODO" as const,
      dataSubject: "",
      contact: "",
      details: "",
      deadline: "",
      actionTaken: "",
    },
  });

  const onSubmit = async (formData: FormData) => {
    try {
      await createRequest({
        organizationId,
        requestType: formData.requestType,
        requestState: formData.requestState,
        dataSubject: formData.dataSubject || undefined,
        contact: formData.contact || undefined,
        details: formData.details || undefined,
        deadline: formatDatetime(formData.deadline),
        actionTaken: formData.actionTaken || undefined,
      });

      toast({
        title: t("createRightsRequestDialog.messages.success"),
        description: t("createRightsRequestDialog.messages.created"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: t("createRightsRequestDialog.messages.error"),
        description: formatError(t("createRightsRequestDialog.errors.create"), error as GraphQLError),
        variant: "error",
      });
    }
  };

  const typeOptions = ["ACCESS", "DELETION", "RECTIFICATION", "PORTABILITY", "OBJECTION", "COMPLAINT"] as const;
  const stateOptions = ["TODO", "IN_PROGRESS", "DONE", "REJECTED"] as const;

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[t("createRightsRequestDialog.breadcrumb.requests"), t("createRightsRequestDialog.breadcrumb.create")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Controller
              control={control}
              name="requestType"
              render={({ field }) => (
                <div>
                  <Label>
                    {t("rightsRequestDetailsPage.fields.requestType")}
                    {" "}
                    *
                  </Label>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    {typeOptions.map(option => (
                      <Option key={option} value={option}>
                        {t(`rightsRequestDetailsPage.types.${option.toLowerCase()}`)}
                      </Option>
                    ))}
                  </Select>
                  {formState.errors.requestType?.message && (
                    <div className="text-red-500 text-sm mt-1">
                      {formState.errors.requestType.message}
                    </div>
                  )}
                </div>
              )}
            />

            <Controller
              control={control}
              name="requestState"
              render={({ field }) => (
                <div>
                  <Label>
                    {t("rightsRequestDetailsPage.fields.state")}
                    {" "}
                    *
                  </Label>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    {stateOptions.map(option => (
                      <Option key={option} value={option}>
                        {t(`rightsRequestDetailsPage.states.${option.toLowerCase()}`)}
                      </Option>
                    ))}
                  </Select>
                  {formState.errors.requestState?.message && (
                    <div className="text-red-500 text-sm mt-1">
                      {formState.errors.requestState.message}
                    </div>
                  )}
                </div>
              )}
            />
          </div>

          <Field
            label={t("rightsRequestDetailsPage.fields.dataSubject")}
            {...register("dataSubject")}
            placeholder={t("createRightsRequestDialog.placeholders.dataSubject")}
            error={formState.errors.dataSubject?.message}
          />

          <Field
            label={t("rightsRequestDetailsPage.fields.contact")}
            {...register("contact")}
            placeholder={t("createRightsRequestDialog.placeholders.contact")}
            error={formState.errors.contact?.message}
          />

          <div>
            <Label>{t("rightsRequestDetailsPage.fields.details")}</Label>
            <Textarea
              {...register("details")}
              placeholder={t("rightsRequestDetailsPage.placeholders.details")}
              rows={3}
            />
            {formState.errors.details?.message && (
              <div className="text-red-500 text-sm mt-1">
                {formState.errors.details.message}
              </div>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>{t("rightsRequestDetailsPage.fields.deadline")}</Label>
              <Input
                type="date"
                {...register("deadline")}
              />
              {formState.errors.deadline?.message && (
                <div className="text-red-500 text-sm mt-1">
                  {formState.errors.deadline.message}
                </div>
              )}
            </div>
          </div>

          <div>
            <Label>{t("rightsRequestDetailsPage.fields.actionTaken")}</Label>
            <Textarea
              {...register("actionTaken")}
              placeholder={t("rightsRequestDetailsPage.placeholders.actionTaken")}
              rows={3}
            />
            {formState.errors.actionTaken?.message && (
              <div className="text-red-500 text-sm mt-1">
                {formState.errors.actionTaken.message}
              </div>
            )}
          </div>
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            variant="primary"
            disabled={formState.isSubmitting}
          >
            {formState.isSubmitting ? t("createRightsRequestDialog.actions.creating") : t("createRightsRequestDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
