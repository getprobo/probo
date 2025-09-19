import { useTranslate } from "@probo/i18n";
import { safeOpenUrl } from "@probo/helpers";
import {
  Avatar,
  Button,
  Card,
  IconPlusLarge,
  IconTrashCan,
  IconPencil,
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
          variant="secondary"
          icon={IconPlusLarge}
          onClick={handleCreate}
        >
          {__("Add Reference")}
        </Button>
      </div>

      <Card padded>
        {references.length === 0 ? (
          <div className="text-center py-12">
            <div className="mx-auto w-12 h-12 bg-tertiary rounded-lg flex items-center justify-center mb-4">
              <IconPlusLarge size={24} className="text-txt-tertiary" />
            </div>
            <h3 className="text-lg font-medium text-txt-primary mb-2">
              {__("No references yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Add customer testimonials and partner references to build trust")}
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {references.map((reference: Reference) => (
              <ReferenceRow
                key={reference.id}
                reference={reference}
                onEdit={() => handleEdit(reference)}
                connectionId={referencesConnectionId}
                onVisitWebsite={() => handleVisitWebsite(reference.websiteUrl)}
              />
            ))}
          </div>
        )}
      </Card>

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
    <div className="flex items-center justify-between p-4 bg-level-1 rounded-lg">
      <div className="flex items-center space-x-4 flex-1">
        <Avatar
          src={reference.logoUrl}
          name={reference.name}
          size="l"
        />

        <div className="flex-1 min-w-0">
          <div className="flex items-center space-x-2 mb-1">
            <button
              type="button"
              onClick={onVisitWebsite}
              className="font-medium text-txt-primary truncate hover:text-primary hover:underline text-left cursor-pointer"
            >
              {reference.name}
            </button>
          </div>

          <p className="text-sm text-txt-secondary line-clamp-2 mb-2">
            {reference.description}
          </p>
        </div>
      </div>

      <div className="flex items-center justify-center space-x-2 ml-4">
        <Button
          variant="tertiary"
          icon={IconPencil}
          onClick={onEdit}
          aria-label="Edit reference"
        />
        <DeleteTrustCenterReferenceDialog
          referenceId={reference.id}
          referenceName={reference.name}
          connectionId={connectionId}
        >
          <Button
            variant="danger"
            icon={IconTrashCan}
            aria-label="Delete reference"
          />
        </DeleteTrustCenterReferenceDialog>
      </div>
    </div>
  );
}
