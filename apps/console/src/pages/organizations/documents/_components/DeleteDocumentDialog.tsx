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

import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, useImperativeHandle } from "react";
import { useTranslation } from "react-i18next";
import { ConnectionHandler, type DataID } from "relay-runtime";

import {
  useBulkDeleteDocumentsMutation,
  useDeleteDocumentMutation,
} from "#/hooks/graph/DocumentGraph";

export type DeleteDocumentDialogRef = {
  open: () => void;
  close: () => void;
};

type DeleteDocumentDialogProps = {
  documentId: string;
  documentTitle: string;
  connections: DataID[];
  onSuccess?: () => void;
};

export const DeleteDocumentDialog = forwardRef<
  DeleteDocumentDialogRef,
  DeleteDocumentDialogProps
>(({ documentId, documentTitle, connections, onSuccess }, ref) => {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [deleteDocument, isDeleting] = useDeleteDocumentMutation();

  useImperativeHandle(ref, () => ({
    open: () => dialogRef.current?.open(),
    close: () => dialogRef.current?.close(),
  }));

  const handleDelete = async () => {
    try {
      await deleteDocument({
        variables: {
          input: { documentId },
          connections,
        },
      });
      dialogRef.current?.close();
      onSuccess?.();
    } catch {
      // The mutation helper already displays the error toast.
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      title={t("deleteDocumentDialog.title.single")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {t("deleteDocumentDialog.confirmation.single", { title: documentTitle })}
        </p>
        <p className="text-txt-secondary mt-2">
          {t("deleteDocumentDialog.irreversible")}
        </p>
      </DialogContent>
      <DialogFooter>
        <Button
          variant="danger"
          onClick={() => void handleDelete()}
          disabled={isDeleting}
          icon={isDeleting ? Spinner : undefined}
        >
          {t("deleteDocumentDialog.actions.delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
});

DeleteDocumentDialog.displayName = "DeleteDocumentDialog";

type DeleteDocumentsDialogProps = {
  documentIds: string[];
  connectionId: DataID;
  onSuccess?: () => void;
};

export const DeleteDocumentsDialog = forwardRef<
  DeleteDocumentDialogRef,
  DeleteDocumentsDialogProps
>(({ documentIds, connectionId, onSuccess }, ref) => {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [deleteDocuments, isDeleting] = useBulkDeleteDocumentsMutation();
  const documentCount = documentIds.length;

  useImperativeHandle(ref, () => ({
    open: () => dialogRef.current?.open(),
    close: () => dialogRef.current?.close(),
  }));

  const handleDelete = async () => {
    try {
      await deleteDocuments({
        variables: { input: { documentIds } },
        updater: (store) => {
          const conn = store.get(connectionId);
          if (conn) {
            documentIds.forEach(id => ConnectionHandler.deleteNode(conn, id));
          }
        },
      });
      dialogRef.current?.close();
      onSuccess?.();
    } catch {
      // The mutation helper already displays the error toast.
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      title={t("deleteDocumentDialog.title.multiple")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {t("deleteDocumentDialog.confirmation.selected", {
            count: documentCount,
          })}
        </p>
        <p className="text-txt-secondary mt-2">
          {t("deleteDocumentDialog.irreversible")}
        </p>
      </DialogContent>
      <DialogFooter>
        <Button
          variant="danger"
          onClick={() => void handleDelete()}
          disabled={isDeleting || documentCount === 0}
          icon={isDeleting ? Spinner : undefined}
        >
          {t("deleteDocumentDialog.actions.delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
});

DeleteDocumentsDialog.displayName = "DeleteDocumentsDialog";
