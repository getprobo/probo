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

import { Button, Dialog, DialogContent, DialogFooter, Field, Label, Option, Spinner, Textarea, useDialogRef } from "@probo/ui";
import { forwardRef, useImperativeHandle, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageCommitmentDialogCreateMutation, CompliancePortalCommitmentIcon } from "#/__generated__/core/CompliancePageCommitmentDialogCreateMutation.graphql";
import type { CompliancePageCommitmentDialogUpdateMutation } from "#/__generated__/core/CompliancePageCommitmentDialogUpdateMutation.graphql";
import type { CompliancePageCommitmentListItemFragment$data } from "#/__generated__/core/CompliancePageCommitmentListItemFragment.graphql";
import { ControlledSelect } from "#/components/form/ControlledField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { COMMITMENT_ICON_VALUES } from "../_lib/commitmentIcons";

const createCommitmentMutation = graphql`
  mutation CompliancePageCommitmentDialogCreateMutation(
    $input: CreateCompliancePortalCommitmentInput!
  ) {
    createCompliancePortalCommitment(input: $input) {
      compliancePortalCommitmentEdge {
        node {
          id
          icon
          eyebrow
          title
          description
          rank
        }
      }
    }
  }
`;

const updateCommitmentMutation = graphql`
  mutation CompliancePageCommitmentDialogUpdateMutation(
    $input: UpdateCompliancePortalCommitmentInput!
  ) {
    updateCompliancePortalCommitment(input: $input) {
      compliancePortalCommitment {
        id
        icon
        eyebrow
        title
        description
        rank
      }
    }
  }
`;

type CommitmentFormData = {
  icon: string;
  eyebrow: string;
  title: string;
  description: string;
};

export type CompliancePageCommitmentDialogRef = {
  openCreate: (groupId: string) => void;
  openEdit: (commitment: CompliancePageCommitmentListItemFragment$data) => void;
};

export const CompliancePageCommitmentDialog = forwardRef<
  CompliancePageCommitmentDialogRef,
  { onChanged: () => void }
>(function CompliancePageCommitmentDialog({ onChanged }, ref) {
  const { t } = useTranslation("organizations/compliance-page");
  const dialogRef = useDialogRef();
  const commitmentSchema = z.object({
    icon: z.string().min(1, t("commitmentDialog.validation.iconRequired")),
    eyebrow: z.string(),
    title: z.string().min(1, t("commitmentDialog.validation.titleRequired")),
    description: z.string().min(
      1,
      t("commitmentDialog.validation.descriptionRequired"),
    ),
  });
  const [mode, setMode] = useState<"create" | "edit">("create");
  const [groupId, setGroupId] = useState<string>("");
  const [commitmentId, setCommitmentId] = useState<string>("");

  const [createCommitment, isCreating] = useMutationWithToasts<CompliancePageCommitmentDialogCreateMutation>(
    createCommitmentMutation,
    { successMessage: t("commitmentDialog.messages.created"), errorMessage: t("commitmentDialog.errors.create") },
  );
  const [updateCommitment, isUpdating] = useMutationWithToasts<CompliancePageCommitmentDialogUpdateMutation>(
    updateCommitmentMutation,
    { successMessage: t("commitmentDialog.messages.updated"), errorMessage: t("commitmentDialog.errors.update") },
  );

  const { register, handleSubmit, control, formState: { errors }, reset } = useFormWithSchema(commitmentSchema, {
    defaultValues: { icon: COMMITMENT_ICON_VALUES[0], eyebrow: "", title: "", description: "" },
  });

  useImperativeHandle(ref, () => ({
    openCreate: (gId: string) => {
      setMode("create");
      setGroupId(gId);
      reset({ icon: COMMITMENT_ICON_VALUES[0], eyebrow: "", title: "", description: "" });
      dialogRef.current?.open();
    },
    openEdit: (commitment) => {
      setMode("edit");
      setCommitmentId(commitment.id);
      reset({
        icon: commitment.icon,
        eyebrow: commitment.eyebrow,
        title: commitment.title,
        description: commitment.description,
      });
      dialogRef.current?.open();
    },
  }));

  const onSubmit = async (data: CommitmentFormData) => {
    const icon = data.icon as CompliancePortalCommitmentIcon;

    if (mode === "create") {
      await createCommitment({
        variables: {
          input: { groupId, icon, eyebrow: data.eyebrow, title: data.title, description: data.description },
        },
        onSuccess: () => {
          reset();
          dialogRef.current?.close();
          onChanged();
        },
      });
    } else {
      await updateCommitment({
        variables: {
          input: { id: commitmentId, icon, eyebrow: data.eyebrow, title: data.title, description: data.description },
        },
        onSuccess: () => {
          reset();
          dialogRef.current?.close();
          onChanged();
        },
      });
    }
  };

  const isSubmitting = isCreating || isUpdating;
  const title = mode === "create"
    ? t("commitmentDialog.title.create")
    : t("commitmentDialog.title.edit");

  return (
    <Dialog ref={dialogRef} title={title} className="max-w-2xl" onClose={() => reset()}>
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <div className="space-y-1.5">
            <Label>{t("commitmentDialog.fields.icon")}</Label>
            <ControlledSelect
              control={control}
              name="icon"
              placeholder={t("commitmentDialog.fields.iconPlaceholder")}
            >
              {COMMITMENT_ICON_VALUES.map(value => (
                <Option key={value} value={value}>
                  {t(`commitmentDialog.icons.${value.toLowerCase()}`)}
                </Option>
              ))}
            </ControlledSelect>
          </div>

          <Field
            {...register("eyebrow")}
            label={t("commitmentDialog.fields.eyebrow")}
            type="text"
            error={errors.eyebrow?.message}
            placeholder={t("commitmentDialog.fields.eyebrowPlaceholder")}
          />

          <Field
            {...register("title")}
            label={t("commitmentDialog.fields.title")}
            type="text"
            required
            error={errors.title?.message}
            placeholder={t("commitmentDialog.fields.titlePlaceholder")}
          />

          <Field
            label={t("commitmentDialog.fields.description")}
            error={errors.description?.message}
            required
          >
            <Textarea
              {...register("description")}
              placeholder={t("commitmentDialog.fields.descriptionPlaceholder")}
              rows={3}
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isSubmitting} icon={isSubmitting ? Spinner : undefined}>
            {mode === "create"
              ? t("commitmentDialog.actions.add")
              : t("commitmentDialog.actions.update")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
});
