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

import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Spinner } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
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

  const { t } = useTranslation("organizations/compliance-page");

  const [deleteFile, isDeleting] = useMutation<DeleteCompliancePageFileDialogMutation>(
    deleteCompliancePageFileMutation,
    {
      successMessage: t("deleteFileDialog.messages.deleted"),
      errorToast: t("deleteFileDialog.errors.delete"),
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
    <Dialog ref={ref} title={t("deleteFileDialog.title")}>
      <DialogContent padded>
        <p>
          {t("deleteFileDialog.description")}
        </p>
      </DialogContent>
      <DialogFooter>
        <Button
          variant="danger"
          onClick={() => void handleDelete()}
          disabled={isDeleting}
        >
          {isDeleting && <Spinner />}
          {t("deleteFileDialog.actions.delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
