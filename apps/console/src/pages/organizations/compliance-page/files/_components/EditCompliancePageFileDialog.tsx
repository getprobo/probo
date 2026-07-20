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
import { Button, Dialog, DialogContent, DialogFooter, Field, Spinner } from "@probo/ui";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageFileListItem_fileFragment$data } from "#/__generated__/core/CompliancePageFileListItem_fileFragment.graphql";
import type { EditCompliancePageFileDialogMutation } from "#/__generated__/core/EditCompliancePageFileDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutation } from "#/lib/relay/useMutation";

const updateCompliancePageFileMutation = graphql`
  mutation EditCompliancePageFileDialogMutation($input: UpdateCompliancePortalFileInput!) {
    updateCompliancePortalFile(input: $input) {
      compliancePortalFile {
        ...CompliancePageFileListItem_fileFragment
      }
    }
  }
`;

export function EditCompliancePageFileDialog(props: {
  file: CompliancePageFileListItem_fileFragment$data;
  onClose: () => void;
}) {
  const { file, onClose } = props;

  const { __ } = useTranslate();

  const editSchema = z.object({
    name: z.string().min(1, __("Name is required")),
    category: z.string().min(1, __("Category is required")),
  });
  const editForm = useFormWithSchema(editSchema, {
    defaultValues: { name: file.name, category: file.category },
  });

  const [updateFile, isUpdating] = useMutation<EditCompliancePageFileDialogMutation>(
    updateCompliancePageFileMutation,
    {
      successMessage: "File updated successfully",
      errorToast: "Failed to update file",
    },
  );

  const handleUpdate = async (data: z.infer<typeof editSchema>) => {
    await updateFile({
      variables: {
        input: {
          id: file.id,
          name: data.name,
          category: data.category,
        },
      },
    });

    onClose();
  };

  return (
    <Dialog defaultOpen={true} title={__("Edit File")} onClose={onClose}>
      <form onSubmit={e => void editForm.handleSubmit(handleUpdate)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            type="text"
            {...editForm.register("name")}
            error={editForm.formState.errors.name?.message}
          />
          <Field
            label={__("Category")}
            type="text"
            {...editForm.register("category")}
            error={editForm.formState.errors.category?.message}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isUpdating}>
            {isUpdating && <Spinner />}
            {__("Save")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
