import { useTranslate } from "@probo/i18n";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  Button,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode } from "react";
import { useFragment, graphql } from "react-relay";
import type { TrustCenterDocumentsCardFragment$key } from "./__generated__/TrustCenterDocumentsCardFragment.graphql";

const trustCenterDocumentDialogFragment = graphql`
  fragment TrustCenterDocumentDialogFragment on Document {
    id
    title
    showOnTrustCenter
  }
`;

type Props = {
  children: ReactNode;
  documents: (TrustCenterDocumentsCardFragment$key & { id: string })[];
  onToggleVisibility: (documentId: string, showOnTrustCenter: boolean) => void;
  disabled?: boolean;
};

export default function TrustCenterDocumentDialog({
  children,
  documents,
  onToggleVisibility,
  disabled,
}: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();

  return (
    <>
      <span onClick={() => !disabled && dialogRef.current?.open()}>
        {children}
      </span>
      <Dialog ref={dialogRef}>
        <DialogContent>
          <div className="text-lg font-semibold mb-4">
            {__("Manage Document Visibility on Trust Center")}
          </div>
          <div className="py-4">
            <p className="text-txt-secondary mb-4">
              {__("Control which documents are visible on your public trust center.")}
            </p>
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {documents.map((documentKey) => (
                <DocumentDialogRow
                  key={documentKey.id}
                  document={documentKey}
                  onToggleVisibility={onToggleVisibility}
                  disabled={disabled}
                />
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => dialogRef.current?.close()}>
              {__("Close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

function DocumentDialogRow({
  document: documentKey,
  onToggleVisibility,
  disabled,
}: {
  document: TrustCenterDocumentsCardFragment$key & { id: string };
  onToggleVisibility: (documentId: string, showOnTrustCenter: boolean) => void;
  disabled?: boolean;
}) {
  const document = useFragment(trustCenterDocumentDialogFragment, documentKey);
  const { __ } = useTranslate();

  return (
    <div className="flex items-center justify-between p-3 border border-border-solid rounded">
      <div className="flex-1">
        <div className="font-medium">{document.title}</div>
      </div>
      <Button
        variant="secondary"
        onClick={() => onToggleVisibility(document.id, !document.showOnTrustCenter)}
        disabled={disabled}
      >
        {document.showOnTrustCenter ? __("Hide") : __("Show")}
      </Button>
    </div>
  );
}
