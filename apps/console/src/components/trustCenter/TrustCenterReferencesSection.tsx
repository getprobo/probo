import { useTranslate } from "@probo/i18n";
import { safeOpenUrl } from "@probo/helpers";
import {
  Avatar,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  IconPlusLarge,
  IconTrashCan,
  IconPencil,
  IconArrowLink,
} from "@probo/ui";
import { type ReactNode, useRef, useState } from "react";
import {
  useTrustCenterReferences,
  useUpdateTrustCenterReferenceRankMutation,
} from "/hooks/graph/TrustCenterReferenceGraph";
import { TrustCenterReferenceDialog, type TrustCenterReferenceDialogRef } from "./TrustCenterReferenceDialog";
import { DeleteTrustCenterReferenceDialog } from "./DeleteTrustCenterReferenceDialog";
import { IfAuthorized } from "/permissions/IfAuthorized";

type Props = {
  trustCenterId: string;
  children?: ReactNode;
};

type Reference = {
  id: string;
  name: string;
  description?: string | null;
  websiteUrl: string;
  logoUrl: string;
  rank: number;
  createdAt: string;
  updatedAt: string;
};

export function TrustCenterReferencesSection({ trustCenterId }: Props) {
  const { __ } = useTranslate();
  const dialogRef = useRef<TrustCenterReferenceDialogRef>(null);
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
  const [refetchKey, setRefetchKey] = useState(0);
  const data = useTrustCenterReferences(trustCenterId, refetchKey);
  const [updateRank] = useUpdateTrustCenterReferenceRankMutation();

  const trustCenterNode = data?.node;
  const references = trustCenterNode?.references?.edges?.map((edge) => edge.node) || [];
  const referencesConnectionId = trustCenterNode?.references?.__id || "";

  const handleCreate = () => {
    if (referencesConnectionId) {
      dialogRef.current?.openCreate(trustCenterId, referencesConnectionId);
    }
  };

  const handleEdit = (reference: Reference) => {
    dialogRef.current?.openEdit(reference);
  };

  const handleVisitWebsite = (websiteUrl: string) => {
    safeOpenUrl(websiteUrl);
  };

  const handleDragStart = (index: number) => {
    setDraggedIndex(index);
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    if (draggedIndex !== index) {
      setDragOverIndex(index);
    }
  };

  const handleDrop = (targetIndex: number) => {
    if (draggedIndex === null || draggedIndex === targetIndex) {
      setDraggedIndex(null);
      setDragOverIndex(null);
      return;
    }

    const draggedRef = references[draggedIndex];
    const targetRank = references[targetIndex].rank;

    updateRank({
      variables: {
        input: {
          id: draggedRef.id,
          rank: targetRank,
        },
      },
      onCompleted: () => {
        setRefetchKey((prev) => prev + 1);
      },
    });

    setDraggedIndex(null);
    setDragOverIndex(null);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-base font-medium">{__("Trusted by")}</h2>
          <p className="text-sm text-txt-tertiary">
            {__("Showcase your customers and partners on your trust center")}
          </p>
        </div>
        <IfAuthorized entity="TrustCenter" action="update">
          <Button
            icon={IconPlusLarge}
            onClick={handleCreate}
          >
            {__("Add Reference")}
          </Button>
        </IfAuthorized>
      </div>

      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Description")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {references.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {__("No references available")}
              </Td>
            </Tr>
          )}
          {references.map((reference: Reference, index: number) => (
            <ReferenceRow
              key={reference.id}
              reference={reference}
              index={index}
              isDragging={draggedIndex === index}
              isDropTarget={dragOverIndex === index && draggedIndex !== index}
              onEdit={() => handleEdit(reference)}
              connectionId={referencesConnectionId}
              onVisitWebsite={() => handleVisitWebsite(reference.websiteUrl)}
              onDragStart={() => handleDragStart(index)}
              onDragOver={(e) => handleDragOver(e, index)}
              onDrop={() => handleDrop(index)}
            />
          ))}
        </Tbody>
      </Table>

      <p className="text-xs text-txt-tertiary">
        {__("Drag and drop references to change their displayed order")}
      </p>

      <TrustCenterReferenceDialog ref={dialogRef} />
    </div>
  );
}

type ReferenceRowProps = {
  reference: Reference;
  index: number;
  isDragging: boolean;
  isDropTarget: boolean;
  onEdit: () => void;
  connectionId: string;
  onVisitWebsite: () => void;
  onDragStart: () => void;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: () => void;
};

function ReferenceRow({
  reference,
  isDragging,
  isDropTarget,
  onEdit,
  connectionId,
  onVisitWebsite,
  onDragStart,
  onDragOver,
  onDrop,
}: ReferenceRowProps) {
  const [isMouseDown, setIsMouseDown] = useState(false);

  const className = [
    isDragging && "opacity-50 cursor-grabbing",
    !isDragging && !isMouseDown && "cursor-grab",
    !isDragging && isMouseDown && "cursor-grabbing",
    isDropTarget && "!bg-primary-50 border-y-2 border-primary-500",
  ].filter(Boolean).join(" ");

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
          <Avatar
            src={reference.logoUrl}
            name={reference.name}
            size="m"
          />
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
            onClick={onVisitWebsite}
          />
          <IfAuthorized entity="TrustCenter" action="update">
            <Button
              variant="secondary"
              icon={IconPencil}
              onClick={onEdit}
            />
          </IfAuthorized>
          <IfAuthorized entity="TrustCenter" action="delete">
            <DeleteTrustCenterReferenceDialog
              referenceId={reference.id}
              referenceName={reference.name}
              connectionId={connectionId}
            >
              <Button
                variant="danger"
                icon={IconTrashCan}
              />
            </DeleteTrustCenterReferenceDialog>
          </IfAuthorized>
        </div>
      </Td>
    </Tr>
  );
}
