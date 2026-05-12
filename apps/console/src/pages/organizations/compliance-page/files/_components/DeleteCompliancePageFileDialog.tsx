// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
