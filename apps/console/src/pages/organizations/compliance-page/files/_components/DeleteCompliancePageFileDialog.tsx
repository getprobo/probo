import { useTranslate } from "@probo/i18n";
import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Spinner } from "@probo/ui";
import { useCallback } from "react";
import { type DataID, graphql } from "relay-runtime";

import type { DeleteCompliancePageFileDialogMutation } from "#/__generated__/core/DeleteCompliancePageFileDialogMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const deleteCompliancePageFileMutation = graphql`
  mutation DeleteCompliancePageFileDialogMutation(
    $input: DeleteTrustCenterFileInput!
    $connections: [ID!]!
  ) {
    deleteTrustCenterFile(input: $input) {
      deletedTrustCenterFileId @deleteEdge(connections: $connections)
    }
  }
`;

export function DeleteCompliancePageFileDialog(props: {
  connectionId: DataID;
  fileId: string | null;
  ref: DialogRef;
  onDelete: () => void;
}) {
  const { connectionId, fileId, ref, onDelete } = props;

  const { __ } = useTranslate();

  const [deleteFile, isDeleting] = useMutationWithToasts<DeleteCompliancePageFileDialogMutation>(
    deleteCompliancePageFileMutation,
    {
      successMessage: "File deleted successfully",
      errorMessage: "Failed to delete file",
    },
  );

  const handleDelete = useCallback(async () => {
    if (!fileId) {
      return;
    }

    await deleteFile({
      variables: {
        input: { id: fileId },
        connections: connectionId ? [connectionId] : [],
      },
      onSuccess: () => {
        ref.current?.close();
        onDelete();
      },
    });
  }, [fileId, deleteFile, ref, connectionId, onDelete]);

  return (
    <Dialog ref={ref} title={__("Delete File")}>
      <DialogContent padded>
        <p>
          {__(
            "Are you sure you want to delete this file? This action cannot be undone.",
          )}
        </p>
      </DialogContent>
      <DialogFooter>
        <Button
          variant="danger"
          onClick={() => void handleDelete()}
          disabled={isDeleting}
        >
          {isDeleting && <Spinner />}
          {__("Delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
