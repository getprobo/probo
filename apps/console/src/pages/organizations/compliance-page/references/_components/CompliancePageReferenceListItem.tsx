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

import { safeOpenUrl } from "@probo/helpers";
import { Avatar, Button, IconArrowLink, IconPencil, IconTrashCan, Td, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { CompliancePageReferenceListItemFragment$data, CompliancePageReferenceListItemFragment$key } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import { DeleteCompliancePageReferenceDialog } from "#/components/compliancePage/DeleteCompliancePageReferenceDialog";

const fragment = graphql`
  fragment CompliancePageReferenceListItemFragment on TrustCenterReference {
    id
    logo {
      downloadUrl
    }
    name
    description
    websiteUrl
    canUpdate: permission(action: "compliance-portal:portal-reference:update")
    canDelete: permission(action: "compliance-portal:portal-reference:delete")
  }
`;

export function CompliancePageReferenceListItem(props: {
  fragmentRef: CompliancePageReferenceListItemFragment$key;
  index: number;
  isDragging: boolean;
  isDropTarget: boolean;
  onEdit: (r: CompliancePageReferenceListItemFragment$data) => void;
  connectionId: DataID;
  onDragStart: () => void;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: () => void;
}) {
  const {
    connectionId,
    fragmentRef,
    isDragging,
    isDropTarget,
    onEdit,
    onDragStart,
    onDragOver,
    onDrop,
  } = props;

  const reference = useFragment<CompliancePageReferenceListItemFragment$key>(fragment, fragmentRef);

  const [isMouseDown, setIsMouseDown] = useState(false);

  const className = [
    isDragging && "opacity-50 cursor-grabbing",
    !isDragging && !isMouseDown && "cursor-grab",
    !isDragging && isMouseDown && "cursor-grabbing",
    isDropTarget && "!bg-primary-50 border-y-2 border-primary-500",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <Tr
      draggable
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDrop={onDrop}
      onMouseDown={() => setIsMouseDown(true)}
      onMouseUp={() => setIsMouseDown(false)}
      onMouseLeave={() => setIsMouseDown(false)}
      className={className}
    >
      <Td>
        <div className="flex items-center gap-3">
          <Avatar src={reference.logo?.downloadUrl} name={reference.name} size="m" />
          <span className="font-medium">{reference.name}</span>
        </div>
      </Td>
      <Td>
        <span className="text-txt-secondary line-clamp-2">
          {reference.description}
        </span>
      </Td>
      <Td noLink width={200} className="text-end">
        <div className="flex gap-2 justify-end">
          <Button
            variant="secondary"
            icon={IconArrowLink}
            onClick={() => safeOpenUrl(reference.websiteUrl)}
          />
          {reference.canUpdate && (
            <Button variant="secondary" icon={IconPencil} onClick={() => onEdit(reference)} />
          )}
          {reference.canDelete && (
            <DeleteCompliancePageReferenceDialog
              referenceId={reference.id}
              referenceName={reference.name}
              connectionId={connectionId}
            >
              <Button variant="danger" icon={IconTrashCan} />
            </DeleteCompliancePageReferenceDialog>
          )}
        </div>
      </Td>
    </Tr>
  );
}
