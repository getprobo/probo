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

import { detectSocialName, safeOpenUrl } from "@probo/helpers";
import {
  Button,
  Card,
  IconArrowLink,
  IconPencil,
  IconTrashCan,
  SocialIcon,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CompliancePageCustomLinkListItem_customLink$key } from "#/__generated__/core/CompliancePageCustomLinkListItem_customLink.graphql";
import type { CompliancePageCustomLinkListItem_deleteMutation } from "#/__generated__/core/CompliancePageCustomLinkListItem_deleteMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const customLinkFragment = graphql`
  fragment CompliancePageCustomLinkListItem_customLink on ComplianceCustomLink {
    id
    name
    url
  }
`;

const deleteMutation = graphql`
  mutation CompliancePageCustomLinkListItem_deleteMutation($input: DeleteComplianceCustomLinkInput!) {
    deleteComplianceCustomLink(input: $input) {
      deletedComplianceCustomLinkId
    }
  }
`;

export interface CompliancePageCustomLinkListItemProps {
  customLinkKey: CompliancePageCustomLinkListItem_customLink$key;
  connectionId: string;
  readOnly: boolean;
  isDragging: boolean;
  isDropTarget: boolean;
  onDragStart: () => void;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: () => void;
  onDragEnd: () => void;
  onEdit: () => void;
}

export function CompliancePageCustomLinkListItem(props: CompliancePageCustomLinkListItemProps) {
  const {
    customLinkKey,
    connectionId,
    readOnly,
    isDragging,
    isDropTarget,
    onDragStart,
    onDragOver,
    onDrop,
    onDragEnd,
    onEdit,
  } = props;

  const { t } = useTranslation("organizations/compliance-page");
  const [isMouseDown, setIsMouseDown] = useState(false);

  const customLink = useFragment(customLinkFragment, customLinkKey);

  const [deleteLink] = useMutation<CompliancePageCustomLinkListItem_deleteMutation>(
    deleteMutation,
    { successMessage: t("externalUrls.messages.deleted"), errorToast: t("externalUrls.errors.delete") },
  );

  const handleDelete = () => {
    void deleteLink({
      variables: { input: { id: customLink.id } },
      updater: (store) => {
        const connection = store.get(connectionId);
        if (!connection) return;
        ConnectionHandler.deleteNode(connection, customLink.id);
      },
    });
  };

  const draggable = !readOnly;

  const className = [
    isDragging && "opacity-50 cursor-grabbing",
    !isDragging && draggable && !isMouseDown && "cursor-grab",
    !isDragging && draggable && isMouseDown && "cursor-grabbing",
    isDropTarget && "ring-2 ring-primary-500",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div
      draggable={draggable}
      onDragStart={draggable ? onDragStart : undefined}
      onDragOver={draggable ? onDragOver : undefined}
      onDrop={draggable ? onDrop : undefined}
      onDragEnd={draggable ? onDragEnd : undefined}
      onMouseDown={draggable ? () => setIsMouseDown(true) : undefined}
      onMouseUp={draggable ? () => setIsMouseDown(false) : undefined}
      onMouseLeave={draggable ? () => setIsMouseDown(false) : undefined}
      className={className}
    >
      <Card padded>
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="min-w-0 space-y-1">
            <div className="flex items-center gap-2">
              <SocialIcon
                socialName={detectSocialName(customLink.url)}
                size={16}
                className="shrink-0"
              />
              <span className="font-medium">{customLink.name}</span>
            </div>
            <p className="truncate text-sm text-txt-secondary">{customLink.url}</p>
          </div>

          <div className="flex shrink-0 items-center gap-2">
            <Button
              variant="secondary"
              icon={IconArrowLink}
              onClick={() => safeOpenUrl(customLink.url)}
              aria-label={t("externalUrls.actions.open")}
            />

            {!readOnly && (
              <>
                <Button
                  variant="secondary"
                  icon={IconPencil}
                  onClick={onEdit}
                  aria-label={t("externalUrls.dialog.titleEdit")}
                />
                <Button
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={handleDelete}
                  aria-label={t("externalUrls.actions.remove")}
                />
              </>
            )}
          </div>
        </div>
      </Card>
    </div>
  );
}
