// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  IconTrashCan,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { useTranslation } from "react-i18next";

import { useDeleteCompliancePageReferenceMutation } from "#/pages/organizations/compliance-page/_lib/compliancePageReferenceMutations";

type Props = {
  children: React.ReactNode;
  referenceId: string;
  referenceName: string;
  connectionId: string;
  onSuccess?: () => void;
};

export function DeleteCompliancePageReferenceDialog({
  children,
  referenceId,
  referenceName,
  connectionId,
  onSuccess,
}: Props) {
  const { t } = useTranslation();
  const ref = useDialogRef();

  const [deleteReference, isDeleting] = useDeleteCompliancePageReferenceMutation();

  const handleDelete = async () => {
    await deleteReference({
      variables: {
        input: {
          id: referenceId,
        },
        connections: [connectionId],
      },
    });

    onSuccess?.();
    ref.current?.close();
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={t("deleteTrustCenterReferenceDialog.title")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {t("deleteTrustCenterReferenceDialog.description", { referenceName })}
        </p>
        <p className="text-txt-secondary mt-2">
          {t("deleteTrustCenterReferenceDialog.warning")}
        </p>
      </DialogContent>

      <DialogFooter>
        <Button
          variant="danger"
          onClick={() => void handleDelete()}
          disabled={isDeleting}
          icon={isDeleting ? Spinner : IconTrashCan}
        >
          {isDeleting ? t("deleteTrustCenterReferenceDialog.actions.deleting") : t("deleteTrustCenterReferenceDialog.actions.delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
