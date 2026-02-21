import { documentClassifications, documentTypes, formatDate, getDocumentClassificationLabel, getDocumentTypeLabel } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Avatar, Badge, Button, Drawer, IconCheckmark1, IconCrossLargeX, IconPencil, PropertyRow } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { DocumentLayoutDrawer_documentFragment$key } from "#/__generated__/core/DocumentLayoutDrawer_documentFragment.graphql";
import type { DocumentLayoutDrawer_versionFragment$key } from "#/__generated__/core/DocumentLayoutDrawer_versionFragment.graphql";
import type { DocumentLayoutDrawerMutation } from "#/__generated__/core/DocumentLayoutDrawerMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { DocumentClassificationOptions } from "#/components/form/DocumentClassificationOptions";
import { DocumentTypeOptions } from "#/components/form/DocumentTypeOptions";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const documentFragment = graphql`
  fragment DocumentLayoutDrawer_documentFragment on Document {
    id
    documentType
    canUpdate: permission(action: "core:document:update")
    approvers(first: 100) {
      edges {
        node {
          id
          fullName
        }
      }
    }
  }
`;

const versionFragment = graphql`
  fragment DocumentLayoutDrawer_versionFragment on DocumentVersion {
    id
    classification
    version
    status
    updatedAt
    publishedAt
  }
`;

const updateDocumentMutation = graphql`
  mutation DocumentLayoutDrawerMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
        documentType
        classification
        approvers(first: 100) {
          edges {
            node {
              id
              fullName
            }
          }
        }
      }
    }
  }
`;

const schema = z.object({
  approverIds: z.array(z.string()).min(1, "At least one approver is required"),
  documentType: z.enum(documentTypes),
  classification: z.enum(documentClassifications),
});

export function DocumentLayoutDrawer(props: {
  documentFragmentRef: DocumentLayoutDrawer_documentFragment$key;
  versionFragmentRef: DocumentLayoutDrawer_versionFragment$key;
}) {
  const { documentFragmentRef, versionFragmentRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const [isEditingApprover, setIsEditingApprover] = useState(false);
  const [isEditingType, setIsEditingType] = useState(false);
  const [isEditingClassification, setIsEditingClassification] = useState(false);

  const document = useFragment<DocumentLayoutDrawer_documentFragment$key>(documentFragment, documentFragmentRef);
  const version = useFragment<DocumentLayoutDrawer_versionFragment$key>(versionFragment, versionFragmentRef);

  const isDraft = version.status === "DRAFT";

  const approvers = document.approvers.edges.map(e => e.node);

  const { control, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        approverIds: approvers.map(a => a.id),
        documentType: document.documentType,
        classification: version.classification,
      },
    },
  );

  const [updateDocument, isUpdatingDocument]
    = useMutationWithToasts<DocumentLayoutDrawerMutation>(
      updateDocumentMutation,
      {
        successMessage: __("Document updated successfully."),
        errorMessage: __("Failed to update document"),
      },
    );

  const handleUpdateApprover = async (data: { approverIds: string[] }) => {
    await updateDocument({
      variables: {
        input: {
          id: document.id,
          approverIds: data.approverIds,
        },
      },
      onSuccess: () => {
        setIsEditingApprover(false);
      },
    });
  };

  const handleUpdateDocumentType = async (data: {
    documentType: (typeof documentTypes)[number];
  }) => {
    await updateDocument({
      variables: {
        input: {
          id: document.id,
          documentType: data.documentType,
        },
      },
      onSuccess: () => {
        setIsEditingType(false);
      },
    });
  };

  const handleUpdateClassification = async (data: {
    classification: (typeof documentClassifications)[number];
  }) => {
    await updateDocument({
      variables: {
        input: {
          id: document.id,
          classification: data.classification,
        },
      },
      onSuccess: () => {
        setIsEditingClassification(false);
      },
    });
  };

  return (
    <Drawer>
      <div className="text-base text-txt-primary font-medium mb-4">
        {__("Properties")}
      </div>
      <PropertyRow label={__("Approvers")}>
        {isEditingApprover
          ? (
              <EditablePropertyContent
                onSave={() => void handleSubmit(handleUpdateApprover)()}
                onCancel={() => {
                  setIsEditingApprover(false);
                  reset();
                }}
                disabled={isUpdatingDocument}
              >
                <PeopleMultiSelectField
                  name="approverIds"
                  control={control}
                  organizationId={organizationId}
                  selectedPeople={approvers}
                  placeholder={__("Add approvers...")}
                />
              </EditablePropertyContent>
            )
          : (
              <ReadOnlyPropertyContent
                onEdit={() => setIsEditingApprover(true)}
                canEdit={document.canUpdate}
              >
                <div className="flex flex-wrap gap-2">
                  {approvers.map(approver => (
                    <Badge key={approver.id} variant="highlight" size="md" className="gap-2">
                      <Avatar name={approver.fullName} />
                      {approver.fullName}
                    </Badge>
                  ))}
                </div>
              </ReadOnlyPropertyContent>
            )}
      </PropertyRow>
      <PropertyRow label={__("Type")}>
        {isEditingType
          ? (
              <EditablePropertyContent
                onSave={() => void handleSubmit(handleUpdateDocumentType)()}
                onCancel={() => {
                  setIsEditingType(false);
                  reset();
                }}
                disabled={isUpdatingDocument}
              >
                <ControlledField
                  name="documentType"
                  control={control}
                  type="select"
                >
                  <DocumentTypeOptions />
                </ControlledField>
              </EditablePropertyContent>
            )
          : (
              <ReadOnlyPropertyContent
                onEdit={() => setIsEditingType(true)}
                canEdit={document.canUpdate}
              >
                <div className="text-sm text-txt-secondary">
                  {getDocumentTypeLabel(__, document.documentType)}
                </div>
              </ReadOnlyPropertyContent>
            )}
      </PropertyRow>
      <PropertyRow label={__("Classification")}>
        {isEditingClassification
          ? (
              <EditablePropertyContent
                onSave={() => void handleSubmit(handleUpdateClassification)()}
                onCancel={() => {
                  setIsEditingClassification(false);
                  reset();
                }}
                disabled={isUpdatingDocument}
              >
                <ControlledField
                  name="classification"
                  control={control}
                  type="select"
                >
                  <DocumentClassificationOptions />
                </ControlledField>
              </EditablePropertyContent>
            )
          : (
              <ReadOnlyPropertyContent
                onEdit={() => setIsEditingClassification(true)}
                canEdit={document.canUpdate}
              >
                <div className="text-sm text-txt-secondary">
                  {getDocumentClassificationLabel(
                    __,
                    version.classification,
                  )}
                </div>
              </ReadOnlyPropertyContent>
            )}
      </PropertyRow>
      <PropertyRow label={__("Status")}>
        <Badge
          variant={isDraft ? "highlight" : "success"}
          size="md"
          className="gap-2"
        >
          {isDraft ? __("Draft") : __("Published")}
        </Badge>
      </PropertyRow>
      <PropertyRow label={__("Version")}>
        <div className="text-sm text-txt-secondary">
          {version.version}
        </div>
      </PropertyRow>
      <PropertyRow label={__("Last modified")}>
        <div className="text-sm text-txt-secondary">
          {formatDate(version.updatedAt)}
        </div>
      </PropertyRow>
      {version.publishedAt && (
        <PropertyRow label={__("Published Date")}>
          <div className="text-sm text-txt-secondary">
            {formatDate(version.publishedAt)}
          </div>
        </PropertyRow>
      )}
    </Drawer>
  );
}

function EditablePropertyContent({
  children,
  onSave,
  onCancel,
  disabled,
}: {
  children: React.ReactNode;
  onSave: () => void;
  onCancel: () => void;
  disabled?: boolean;
}) {
  return (
    <div className="flex items-center gap-2">
      <div className="flex-1">{children}</div>
      <Button
        variant="quaternary"
        icon={IconCheckmark1}
        onClick={onSave}
        disabled={disabled}
      />
      <Button variant="quaternary" icon={IconCrossLargeX} onClick={onCancel} />
    </div>
  );
}

function ReadOnlyPropertyContent({
  children,
  onEdit,
  canEdit = true,
}: {
  children: React.ReactNode;
  onEdit: () => void;
  canEdit?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-3">
      {children}
      {canEdit && (
        <Button variant="quaternary" icon={IconPencil} onClick={onEdit} />
      )}
    </div>
  );
}
