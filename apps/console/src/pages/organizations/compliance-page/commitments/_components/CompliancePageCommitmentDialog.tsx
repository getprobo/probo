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

import { useTranslate } from "@probo/i18n";
import { Button, Dialog, DialogContent, DialogFooter, Field, Label, Option, Spinner, Textarea, useDialogRef } from "@probo/ui";
import { forwardRef, useImperativeHandle, useState } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageCommitmentDialogCreateMutation, CompliancePortalCommitmentIcon } from "#/__generated__/core/CompliancePageCommitmentDialogCreateMutation.graphql";
import type { CompliancePageCommitmentDialogUpdateMutation } from "#/__generated__/core/CompliancePageCommitmentDialogUpdateMutation.graphql";
import type { CompliancePageCommitmentListItemFragment$data } from "#/__generated__/core/CompliancePageCommitmentListItemFragment.graphql";
import { ControlledSelect } from "#/components/form/ControlledField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { COMMITMENT_ICON_LABELS, COMMITMENT_ICON_VALUES } from "../_lib/commitmentIcons";

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

const commitmentSchema = z.object({
  icon: z.string().min(1, "Icon is required"),
  eyebrow: z.string(),
  title: z.string().min(1, "Title is required"),
  description: z.string().min(1, "Description is required"),
});

type CommitmentFormData = z.infer<typeof commitmentSchema>;

export type CompliancePageCommitmentDialogRef = {
  openCreate: (groupId: string) => void;
  openEdit: (commitment: CompliancePageCommitmentListItemFragment$data) => void;
};

export const CompliancePageCommitmentDialog = forwardRef<
  CompliancePageCommitmentDialogRef,
  { onChanged: () => void }
>(function CompliancePageCommitmentDialog({ onChanged }, ref) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [mode, setMode] = useState<"create" | "edit">("create");
  const [groupId, setGroupId] = useState<string>("");
  const [commitmentId, setCommitmentId] = useState<string>("");

  const [createCommitment, isCreating] = useMutationWithToasts<CompliancePageCommitmentDialogCreateMutation>(
    createCommitmentMutation,
    { successMessage: __("Commitment created successfully"), errorMessage: __("Failed to create commitment") },
  );
  const [updateCommitment, isUpdating] = useMutationWithToasts<CompliancePageCommitmentDialogUpdateMutation>(
    updateCommitmentMutation,
    { successMessage: __("Commitment updated successfully"), errorMessage: __("Failed to update commitment") },
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
  const title = mode === "create" ? __("Add Commitment") : __("Edit Commitment");

  return (
    <Dialog ref={dialogRef} title={title} className="max-w-2xl" onClose={() => reset()}>
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <div className="space-y-1.5">
            <Label>{__("Icon")}</Label>
            <ControlledSelect control={control} name="icon" placeholder={__("Select an icon")}>
              {COMMITMENT_ICON_VALUES.map(value => (
                <Option key={value} value={value}>
                  {COMMITMENT_ICON_LABELS[value]}
                </Option>
              ))}
            </ControlledSelect>
          </div>

          <Field
            {...register("eyebrow")}
            label={__("Eyebrow")}
            type="text"
            error={errors.eyebrow?.message}
            placeholder={__("Small accent label above the title")}
          />

          <Field
            {...register("title")}
            label={__("Title")}
            type="text"
            required
            error={errors.title?.message}
            placeholder={__("Commitment headline")}
          />

          <Field label={__("Description")} error={errors.description?.message} required>
            <Textarea
              {...register("description")}
              placeholder={__("Supporting body copy")}
              rows={3}
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isSubmitting} icon={isSubmitting ? Spinner : undefined}>
            {mode === "create" ? __("Add Commitment") : __("Update Commitment")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
});
