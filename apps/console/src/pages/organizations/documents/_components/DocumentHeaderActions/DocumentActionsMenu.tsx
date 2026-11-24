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
  ArchiveIcon,
  ArrowCounterClockwiseIcon,
  DotsThreeVerticalIcon,
  DownloadSimpleIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentActionsMenu_document$key } from "#/__generated__/core/DocumentActionsMenu_document.graphql";

import { useUnarchiveDocument } from "../../_lib/useUnarchiveDocument";

import { DocumentArchiveDialog } from "./DocumentArchiveDialog";
import { DocumentDeleteDialog } from "./DocumentDeleteDialog";
import { DocumentPdfExportForm } from "./DocumentPdfExportForm";

const documentActionsMenuFragment = graphql`
  fragment DocumentActionsMenu_document on Document {
    id
    title
    status
    canArchive: permission(action: "core:document:archive")
    canUnarchive: permission(action: "core:document:unarchive")
    canDelete: permission(action: "core:document:delete")
    publishedVersion {
      ...DocumentPdfExportForm_version
    }
  }
`;

interface DocumentActionsMenuProps {
  documentKey: DocumentActionsMenu_document$key;
}

// Downloading needs a published version. Archiving asks first, while
// unarchiving runs straight from the menu.
export function DocumentActionsMenu({ documentKey }: DocumentActionsMenuProps) {
  const document = useFragment(documentActionsMenuFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const [unarchiveDocument, isUnarchiving] = useUnarchiveDocument();
  const [downloadOpen, setDownloadOpen] = useState(false);
  const [archiveOpen, setArchiveOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const publishedVersion = document.publishedVersion;
  const archived = document.status === "ARCHIVED";

  function handleUnarchive() {
    void unarchiveDocument({
      variables: { input: { documentId: document.id } },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <>
      <Dropdown>
        <Tooltip>
          <TooltipTrigger
            render={(
              <DropdownTrigger
                render={(
                  <IconButton
                    size={1}
                    variant="soft"
                    color="neutral"
                    aria-label={t("details.actions.menu")}
                  />
                )}
              />
            )}
          >
            <DotsThreeVerticalIcon weight="bold" />
          </TooltipTrigger>
          <TooltipPopup>{t("details.actions.menu")}</TooltipPopup>
        </Tooltip>
        <DropdownPopup align="end">
          <DropdownItem
            color="neutral"
            iconStart={<DownloadSimpleIcon />}
            disabled={publishedVersion == null}
            onClick={() => setDownloadOpen(true)}
          >
            {t("details.actions.download.item")}
          </DropdownItem>
          {document.canArchive && !archived && (
            <DropdownItem
              color="neutral"
              iconStart={<ArchiveIcon />}
              onClick={() => setArchiveOpen(true)}
            >
              {t("details.actions.archive.item")}
            </DropdownItem>
          )}
          {document.canUnarchive && archived && (
            <DropdownItem
              color="neutral"
              iconStart={<ArrowCounterClockwiseIcon />}
              disabled={isUnarchiving}
              onClick={handleUnarchive}
            >
              {t("details.actions.unarchive")}
            </DropdownItem>
          )}
          {document.canDelete && (
            <DropdownItem
              color="error"
              iconStart={<TrashIcon />}
              onClick={() => setDeleteOpen(true)}
            >
              {t("details.actions.delete.item")}
            </DropdownItem>
          )}
        </DropdownPopup>
      </Dropdown>
      {publishedVersion != null && (
        <Dialog open={downloadOpen} onOpenChange={setDownloadOpen}>
          <DialogPopup>
            <DocumentPdfExportForm versionKey={publishedVersion} onDone={() => setDownloadOpen(false)} />
          </DialogPopup>
        </Dialog>
      )}
      <DocumentArchiveDialog
        documentId={document.id}
        title={document.title}
        open={archiveOpen}
        onOpenChange={setArchiveOpen}
      />
      <DocumentDeleteDialog
        documentId={document.id}
        title={document.title}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
      />
    </>
  );
}
