import { useTranslate } from "@probo/i18n";
import { Button, Dialog, DialogContent, DialogFooter, Field, Spinner } from "@probo/ui";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageFileListItem_fileFragment$data } from "#/__generated__/core/CompliancePageFileListItem_fileFragment.graphql";
import type { EditCompliancePageFileDialogMutation } from "#/__generated__/core/EditCompliancePageFileDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const updateCompliancePageFileMutation = graphql`
  mutation EditCompliancePageFileDialogMutation($input: UpdateTrustCenterFileInput!) {
    updateTrustCenterFile(input: $input) {
      trustCenterFile {
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

  const [updateFile, isUpdating] = useMutationWithToasts<EditCompliancePageFileDialogMutation>(
    updateCompliancePageFileMutation,
    {
      successMessage: "File updated successfully",
      errorMessage: "Failed to update file",
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
      onSuccess: () => {
        onClose();
      },
    });
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
