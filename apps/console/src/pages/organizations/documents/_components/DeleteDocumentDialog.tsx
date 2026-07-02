// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, useImperativeHandle } from "react";
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
  const { __ } = useTranslate();
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
      title={__("Delete document")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {sprintf(
            __("Are you sure you want to delete the document \"%s\"?"),
            documentTitle,
          )}
        </p>
        <p className="text-txt-secondary mt-2">
          {__("This action cannot be undone.")}
        </p>
      </DialogContent>
      <DialogFooter>
        <Button
          variant="danger"
          onClick={() => void handleDelete()}
          disabled={isDeleting}
          icon={isDeleting ? Spinner : undefined}
        >
          {__("Delete")}
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
  const { __ } = useTranslate();
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
      title={__("Delete documents")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {documentCount === 1
            ? __("Are you sure you want to delete 1 selected document?")
            : sprintf(
                __("Are you sure you want to delete %s selected documents?"),
                documentCount,
              )}
        </p>
        <p className="text-txt-secondary mt-2">
          {__("This action cannot be undone.")}
        </p>
      </DialogContent>
      <DialogFooter>
        <Button
          variant="danger"
          onClick={() => void handleDelete()}
          disabled={isDeleting || documentCount === 0}
          icon={isDeleting ? Spinner : undefined}
        >
          {__("Delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
});

DeleteDocumentsDialog.displayName = "DeleteDocumentsDialog";
