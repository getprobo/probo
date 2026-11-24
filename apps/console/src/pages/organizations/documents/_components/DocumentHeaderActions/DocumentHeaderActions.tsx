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

import { UploadSimpleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useQueryLoader } from "react-relay";

import type { DocumentHeaderActions_document$key } from "#/__generated__/core/DocumentHeaderActions_document.graphql";
import type { DocumentPublishReviewCalloutQuery } from "#/__generated__/core/DocumentPublishReviewCalloutQuery.graphql";

import { useDocumentVersionParam } from "../../_lib/useDocumentVersionParam";
import { documentHeaderActions } from "../../variants";

import { DocumentActionsMenu } from "./DocumentActionsMenu";
import { DocumentPublishForm } from "./DocumentPublishForm";
import { documentPublishReviewCalloutQuery } from "./DocumentPublishReviewCallout";

const documentHeaderActionsFragment = graphql`
  fragment DocumentHeaderActions_document on Document {
    id
    status
    writeMode
    canPublish: permission(action: "core:document-version:publish")
    pendingApprovalCount: versions(first: 0, filter: { statuses: [PENDING_APPROVAL] }) {
      totalCount
    }
    ...DocumentPublishForm_document
    ...DocumentActionsMenu_document
  }
`;

interface DocumentHeaderActionsProps {
  documentKey: DocumentHeaderActions_document$key;
}

// Publishing waits while a version is pending approval or a past version is
// shown, and archived or generated documents cannot be published. The form mounts with the dialog
// popup, so each opening starts from the default approvers.
export function DocumentHeaderActions({ documentKey }: DocumentHeaderActionsProps) {
  const document = useFragment(documentHeaderActionsFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const [publishOpen, setPublishOpen] = useState(false);
  const [reviewQueryRef, loadReview] = useQueryLoader<DocumentPublishReviewCalloutQuery>(
    documentPublishReviewCalloutQuery,
  );
  const { versionId } = useDocumentVersionParam();
  const { root } = documentHeaderActions();
  const canPublish = document.canPublish
    && document.pendingApprovalCount.totalCount === 0
    && document.status !== "ARCHIVED"
    && document.writeMode !== "GENERATED"
    && versionId == null;

  return (
    <div className={root()}>
      {canPublish && (
        <>
          <Button
            size={1}
            variant="solid"
            color="neutral"
            highContrast
            iconStart={<UploadSimpleIcon />}
            onClick={() => {
              loadReview({ documentId: document.id }, { fetchPolicy: "network-only" });
              setPublishOpen(true);
            }}
          >
            {t("details.actions.publish.button")}
          </Button>
          <Dialog open={publishOpen} onOpenChange={setPublishOpen}>
            <DialogPopup>
              <DocumentPublishForm
                documentKey={document}
                reviewQueryRef={reviewQueryRef ?? null}
                onDone={() => setPublishOpen(false)}
              />
            </DialogPopup>
          </Dialog>
        </>
      )}
      <DocumentActionsMenu documentKey={document} />
    </div>
  );
}
