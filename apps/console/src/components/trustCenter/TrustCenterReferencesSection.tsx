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
import { type ReactNode, useRef } from "react";
import {
  useTrustCenterReferences,
} from "/hooks/graph/TrustCenterReferenceGraph";
import { TrustCenterReferenceDialog, type TrustCenterReferenceDialogRef } from "./TrustCenterReferenceDialog";
import { DeleteTrustCenterReferenceDialog } from "./DeleteTrustCenterReferenceDialog";

type Props = {
  trustCenterId: string;
  children?: ReactNode;
};

type Reference = {
  id: string;
  name: string;
  description: string;
  websiteUrl: string;
  logoUrl: string;
  createdAt: string;
  updatedAt: string;
};

export function TrustCenterReferencesSection({ trustCenterId }: Props) {
  const { __ } = useTranslate();
  const dialogRef = useRef<TrustCenterReferenceDialogRef>(null);
  const data = useTrustCenterReferences(trustCenterId);

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

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-base font-medium">{__("Trusted by")}</h2>
          <p className="text-sm text-txt-tertiary">
            {__("Showcase your customers and partners on your trust center")}
          </p>
        </div>
        <Button
          icon={IconPlusLarge}
          onClick={handleCreate}
        >
          {__("Add Reference")}
        </Button>
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
          {references.map((reference: Reference) => (
            <ReferenceRow
              key={reference.id}
              reference={reference}
              onEdit={() => handleEdit(reference)}
              connectionId={referencesConnectionId}
              onVisitWebsite={() => handleVisitWebsite(reference.websiteUrl)}
            />
          ))}
        </Tbody>
      </Table>

      <TrustCenterReferenceDialog ref={dialogRef} />
    </div>
  );
}

type ReferenceRowProps = {
  reference: Reference;
  onEdit: () => void;
  connectionId: string;
  onVisitWebsite: () => void;
};

function ReferenceRow({ reference, onEdit, connectionId, onVisitWebsite }: ReferenceRowProps) {
  return (
    <Tr>
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
          <Button
            variant="secondary"
            icon={IconPencil}
            onClick={onEdit}
          />
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
        </div>
      </Td>
    </Tr>
  );
}
