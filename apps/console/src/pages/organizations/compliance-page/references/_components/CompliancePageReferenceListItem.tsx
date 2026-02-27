import { safeOpenUrl } from "@probo/helpers";
import { Button, IconArrowLink, IconPencil, IconTrashCan, Td, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

function GripHandle() {
  return (
    <svg width="10" height="14" viewBox="0 0 10 14" fill="currentColor" aria-hidden>
      <circle cx="2.5" cy="2" r="1.5" />
      <circle cx="7.5" cy="2" r="1.5" />
      <circle cx="2.5" cy="7" r="1.5" />
      <circle cx="7.5" cy="7" r="1.5" />
      <circle cx="2.5" cy="12" r="1.5" />
      <circle cx="7.5" cy="12" r="1.5" />
    </svg>
  );
}

import type { CompliancePageReferenceListItemFragment$data, CompliancePageReferenceListItemFragment$key } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import { DeleteTrustCenterReferenceDialog } from "#/components/trustCenter/DeleteTrustCenterReferenceDialog";

const fragment = graphql`
  fragment CompliancePageReferenceListItemFragment on TrustCenterReference {
    id
    logoUrl
    name
    description
    websiteUrl
    canUpdate: permission(action: "core:trust-center-reference:update")
    canDelete: permission(action: "core:trust-center-reference:delete")
  }
`;

export function CompliancePageReferenceListItem(props: {
  fragmentRef: CompliancePageReferenceListItemFragment$key;
  isDragging: boolean;
  onEdit: (r: CompliancePageReferenceListItemFragment$data) => void;
  connectionId: DataID;
  onDragStart: (data: { name: string; description: string | null | undefined; logoUrl: string | null }) => void;
  onDragEnter: (e: React.DragEvent) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDragEnd: () => void;
  onDrop: (e: React.DragEvent) => void;
}) {
  const {
    connectionId,
    fragmentRef,
    isDragging,
    onEdit,
    onDragStart,
    onDragEnter,
    onDragOver,
    onDragEnd,
    onDrop,
  } = props;

  const reference = useFragment<CompliancePageReferenceListItemFragment$key>(fragment, fragmentRef);

  const [isMouseDown, setIsMouseDown] = useState(false);

  const rowClassName = isDragging ? "[visibility:collapse]" : isMouseDown ? "cursor-grabbing" : "cursor-grab";

  return (
    <Tr
      draggable
      onDragStart={() => onDragStart({ name: reference.name, description: reference.description, logoUrl: reference.logoUrl })}
      onDragEnter={onDragEnter}
      onDragOver={onDragOver}
      onDragEnd={onDragEnd}
      onDrop={onDrop}
      onMouseDown={() => setIsMouseDown(true)}
      onMouseUp={() => setIsMouseDown(false)}
      onMouseLeave={() => setIsMouseDown(false)}
      className={rowClassName}
    >
      <Td width={32} noLink className="text-txt-tertiary pr-3">
        <GripHandle />
      </Td>
      <Td>
        <div className="flex items-center gap-4">
          <div className="size-8 flex-shrink-0 rounded bg-white flex items-center justify-center border border-border-low overflow-hidden">
            <img src={reference.logoUrl ?? undefined} alt={reference.name} className="size-7 object-contain" />
          </div>
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
