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

import { Button, Dialog, DialogContent, DialogFooter, Field, Spinner, Textarea, useDialogRef } from "@probo/ui";
import { forwardRef, useImperativeHandle, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageCommitmentGroupDialogCreateMutation } from "#/__generated__/core/CompliancePageCommitmentGroupDialogCreateMutation.graphql";
import type { CompliancePageCommitmentGroupDialogUpdateMutation } from "#/__generated__/core/CompliancePageCommitmentGroupDialogUpdateMutation.graphql";
import type { CompliancePageCommitmentGroupListItemFragment$data } from "#/__generated__/core/CompliancePageCommitmentGroupListItemFragment.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const createGroupMutation = graphql`
  mutation CompliancePageCommitmentGroupDialogCreateMutation(
    $input: CreateCompliancePortalCommitmentGroupInput!
  ) {
    createCompliancePortalCommitmentGroup(input: $input) {
      compliancePortalCommitmentGroupEdge {
        node {
          id
          title
          description
          rank
        }
      }
    }
  }
`;

const updateGroupMutation = graphql`
  mutation CompliancePageCommitmentGroupDialogUpdateMutation(
    $input: UpdateCompliancePortalCommitmentGroupInput!
  ) {
    updateCompliancePortalCommitmentGroup(input: $input) {
      compliancePortalCommitmentGroup {
        id
        title
        description
        rank
      }
    }
  }
`;

type GroupFormData = { title: string; description: string };

export type CompliancePageCommitmentGroupDialogRef = {
  openCreate: (compliancePortalId: string) => void;
  openEdit: (group: CompliancePageCommitmentGroupListItemFragment$data) => void;
};

export const CompliancePageCommitmentGroupDialog = forwardRef<
  CompliancePageCommitmentGroupDialogRef,
  { onChanged: () => void }
>(function CompliancePageCommitmentGroupDialog({ onChanged }, ref) {
  const { t } = useTranslation("organizations/compliance-page");
  const dialogRef = useDialogRef();
  const groupSchema = z.object({
    title: z.string().min(1, t("commitmentGroupDialog.validation.titleRequired")),
    description: z.string().min(
      1,
      t("commitmentGroupDialog.validation.descriptionRequired"),
    ),
  });
  const [mode, setMode] = useState<"create" | "edit">("create");
  const [compliancePortalId, setCompliancePortalId] = useState<string>("");
  const [groupId, setGroupId] = useState<string>("");

  const [createGroup, isCreating] = useMutationWithToasts<CompliancePageCommitmentGroupDialogCreateMutation>(
    createGroupMutation,
    { successMessage: t("commitmentGroupDialog.messages.created"), errorMessage: t("commitmentGroupDialog.errors.create") },
  );
  const [updateGroup, isUpdating] = useMutationWithToasts<CompliancePageCommitmentGroupDialogUpdateMutation>(
    updateGroupMutation,
    { successMessage: t("commitmentGroupDialog.messages.updated"), errorMessage: t("commitmentGroupDialog.errors.update") },
  );

  const { register, handleSubmit, formState: { errors }, reset } = useFormWithSchema(groupSchema, {
    defaultValues: { title: "", description: "" },
  });

  useImperativeHandle(ref, () => ({
    openCreate: (tId: string) => {
      setMode("create");
      setCompliancePortalId(tId);
      reset({ title: "", description: "" });
      dialogRef.current?.open();
    },
    openEdit: (group) => {
      setMode("edit");
      setGroupId(group.id);
      reset({ title: group.title, description: group.description });
      dialogRef.current?.open();
    },
  }));

  const onSubmit = async (data: GroupFormData) => {
    if (mode === "create") {
      await createGroup({
        variables: { input: { compliancePortalId, title: data.title, description: data.description } },
        onSuccess: () => {
          reset();
          dialogRef.current?.close();
          onChanged();
        },
      });
    } else {
      await updateGroup({
        variables: { input: { id: groupId, title: data.title, description: data.description } },
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
    ? t("commitmentGroupDialog.title.create")
    : t("commitmentGroupDialog.title.edit");

  return (
    <Dialog ref={dialogRef} title={title} className="max-w-2xl" onClose={() => reset()}>
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <Field
            {...register("title")}
            label={t("commitmentGroupDialog.fields.title")}
            type="text"
            required
            error={errors.title?.message}
            placeholder={t("commitmentGroupDialog.fields.titlePlaceholder")}
          />
          <Field
            label={t("commitmentGroupDialog.fields.description")}
            error={errors.description?.message}
            required
          >
            <Textarea
              {...register("description")}
              placeholder={t("commitmentGroupDialog.fields.descriptionPlaceholder")}
              rows={3}
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isSubmitting} icon={isSubmitting ? Spinner : undefined}>
            {mode === "create"
              ? t("commitmentGroupDialog.actions.add")
              : t("commitmentGroupDialog.actions.update")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
});
