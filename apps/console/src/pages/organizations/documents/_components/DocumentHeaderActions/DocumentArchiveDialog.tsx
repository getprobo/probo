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

import { ArchiveIcon } from "@phosphor-icons/react";
import { ConfirmDialog } from "@probo/ui/src/v2/Dialog/ConfirmDialog";
import { useTranslation } from "react-i18next";

import { useArchiveDocument } from "../../_lib/useArchiveDocument";

interface DocumentArchiveDialogProps {
  documentId: string;
  title: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DocumentArchiveDialog({ documentId, title, open, onOpenChange }: DocumentArchiveDialogProps) {
  const { t } = useTranslation("organizations/documents");
  const [archiveDocument, isArchiving] = useArchiveDocument();

  function handleArchive() {
    void archiveDocument({
      variables: { input: { documentId } },
    }).then(() => {
      onOpenChange(false);
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("details.actions.archive.title")}
      description={t("details.actions.archive.description", { title })}
      cancelLabel={t("details.actions.cancel")}
      confirmLabel={t("details.actions.archive.confirm")}
      confirmIcon={<ArchiveIcon />}
      confirmColor="amber"
      loading={isArchiving}
      onConfirm={handleArchive}
    />
  );
}
