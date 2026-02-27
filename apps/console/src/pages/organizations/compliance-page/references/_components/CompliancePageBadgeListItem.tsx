import { Button, IconPencil, IconTrashCan, Td, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { CompliancePageBadgeListItemFragment$data, CompliancePageBadgeListItemFragment$key } from "#/__generated__/core/CompliancePageBadgeListItemFragment.graphql";
import { useDeleteComplianceBadgeMutation } from "#/hooks/graph/ComplianceBadgeGraph";

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

const fragment = graphql`
  fragment CompliancePageBadgeListItemFragment on ComplianceBadge {
    id
    iconUrl
    name
    canUpdate: permission(action: "core:compliance-badge:update")
    canDelete: permission(action: "core:compliance-badge:delete")
  }
`;

export function CompliancePageBadgeListItem(props: {
  fragmentRef: CompliancePageBadgeListItemFragment$key;
  isDragging: boolean;
  onEdit: (b: CompliancePageBadgeListItemFragment$data) => void;
  connectionId: DataID;
  onDragStart: (data: { name: string; iconUrl: string }) => void;
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

  const badge = useFragment<CompliancePageBadgeListItemFragment$key>(fragment, fragmentRef);

  const [isMouseDown, setIsMouseDown] = useState(false);
  const [deleteBadge] = useDeleteComplianceBadgeMutation();

  const rowClassName = isDragging ? "[visibility:collapse]" : isMouseDown ? "cursor-grabbing" : "cursor-grab";

  const handleDelete = async () => {
    await deleteBadge({
      variables: {
        input: { id: badge.id },
        connections: [connectionId],
      },
    });
  };

  return (
    <Tr
      draggable
      onDragStart={() => onDragStart({ name: badge.name, iconUrl: badge.iconUrl })}
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
            <img src={badge.iconUrl} alt={badge.name} className="size-7 object-contain" />
          </div>
          <span className="font-medium">{badge.name}</span>
        </div>
      </Td>
      <Td noLink width={200} className="text-end">
        <div className="flex gap-2 justify-end">
          {badge.canUpdate && (
            <Button variant="secondary" icon={IconPencil} onClick={() => onEdit(badge)} />
          )}
          {badge.canDelete && (
            <Button variant="danger" icon={IconTrashCan} onClick={() => void handleDelete()} />
          )}
        </div>
      </Td>
    </Tr>
  );
}
