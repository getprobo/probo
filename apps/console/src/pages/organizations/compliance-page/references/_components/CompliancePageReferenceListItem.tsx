import { safeOpenUrl } from "@probo/helpers";
import { Avatar, Button, IconArrowLink, IconPencil, IconTrashCan, Td, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { CompliancePageReferenceListItemFragment$data, CompliancePageReferenceListItemFragment$key } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import { DeleteTrustCenterReferenceDialog } from "#/components/trustCenter/DeleteTrustCenterReferenceDialog";

const fragment = graphql`
  fragment CompliancePageReferenceListItemFragment on TrustCenterReference {
    id
    logoUrl
    name
    description
    rank
    websiteUrl
    canUpdate: permission(action: "core:trust-center-reference:update")
    canDelete: permission(action: "core:trust-center-reference:delete")
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
          <Avatar src={reference.logoUrl} name={reference.name} size="m" />
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
            <DeleteTrustCenterReferenceDialog
              referenceId={reference.id}
              referenceName={reference.name}
              connectionId={connectionId}
            >
              <Button variant="danger" icon={IconTrashCan} />
            </DeleteTrustCenterReferenceDialog>
          )}
        </div>
      </Td>
    </Tr>
  );
}
