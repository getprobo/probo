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
import { Button, Dialog, DialogContent, DialogFooter, Field, Spinner, Textarea, useDialogRef } from "@probo/ui";
import { forwardRef, useImperativeHandle, useState } from "react";
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

const groupSchema = z.object({
  title: z.string().min(1, "Title is required"),
  description: z.string().min(1, "Description is required"),
});

type GroupFormData = z.infer<typeof groupSchema>;

export type CompliancePageCommitmentGroupDialogRef = {
  openCreate: (compliancePortalId: string) => void;
  openEdit: (group: CompliancePageCommitmentGroupListItemFragment$data) => void;
};

export const CompliancePageCommitmentGroupDialog = forwardRef<
  CompliancePageCommitmentGroupDialogRef,
  { onChanged: () => void }
>(function CompliancePageCommitmentGroupDialog({ onChanged }, ref) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [mode, setMode] = useState<"create" | "edit">("create");
  const [compliancePortalId, setCompliancePortalId] = useState<string>("");
  const [groupId, setGroupId] = useState<string>("");

  const [createGroup, isCreating] = useMutationWithToasts<CompliancePageCommitmentGroupDialogCreateMutation>(
    createGroupMutation,
    { successMessage: __("Group created successfully"), errorMessage: __("Failed to create group") },
  );
  const [updateGroup, isUpdating] = useMutationWithToasts<CompliancePageCommitmentGroupDialogUpdateMutation>(
    updateGroupMutation,
    { successMessage: __("Group updated successfully"), errorMessage: __("Failed to update group") },
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
  const title = mode === "create" ? __("Add Group") : __("Edit Group");

  return (
    <Dialog ref={dialogRef} title={title} className="max-w-2xl" onClose={() => reset()}>
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <Field
            {...register("title")}
            label={__("Title")}
            type="text"
            required
            error={errors.title?.message}
            placeholder={__("e.g. Data Protection")}
          />
          <Field label={__("Description")} error={errors.description?.message} required>
            <Textarea
              {...register("description")}
              placeholder={__("Describe what this group of commitments covers")}
              rows={3}
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isSubmitting} icon={isSubmitting ? Spinner : undefined}>
            {mode === "create" ? __("Add Group") : __("Update Group")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
});
