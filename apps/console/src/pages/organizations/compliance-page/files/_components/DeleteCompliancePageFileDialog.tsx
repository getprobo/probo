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
import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Spinner } from "@probo/ui";
import { useCallback } from "react";
import { type DataID, graphql } from "relay-runtime";

import type { DeleteCompliancePageFileDialogMutation } from "#/__generated__/core/DeleteCompliancePageFileDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const deleteCompliancePageFileMutation = graphql`
  mutation DeleteCompliancePageFileDialogMutation(
    $input: DeleteCompliancePortalFileInput!
    $connections: [ID!]!
  ) {
    deleteCompliancePortalFile(input: $input) {
      deletedCompliancePortalFileId @deleteEdge(connections: $connections)
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

  const [deleteFile, isDeleting] = useMutation<DeleteCompliancePageFileDialogMutation>(
    deleteCompliancePageFileMutation,
    {
      successMessage: "File deleted successfully",
      errorToast: "Failed to delete file",
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
    });

    ref.current?.close();
    onDelete();
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
