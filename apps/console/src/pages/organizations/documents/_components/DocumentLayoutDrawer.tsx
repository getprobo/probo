// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { documentClassifications, documentTypes, formatDate, getDocumentClassificationLabel, getDocumentTypeLabel } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Drawer, IconCheckmark1, IconCrossLargeX, IconPencil, PropertyRow, useToast } from "@probo/ui";
import { useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { DocumentLayoutDrawer_documentFragment$key } from "#/__generated__/core/DocumentLayoutDrawer_documentFragment.graphql";
import type { DocumentLayoutDrawer_updateClassificationMutation } from "#/__generated__/core/DocumentLayoutDrawer_updateClassificationMutation.graphql";
import type { DocumentLayoutDrawer_versionFragment$key } from "#/__generated__/core/DocumentLayoutDrawer_versionFragment.graphql";
import type { DocumentLayoutDrawerMutation } from "#/__generated__/core/DocumentLayoutDrawerMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { DocumentClassificationOptions } from "#/components/form/DocumentClassificationOptions";
import { DocumentTypeOptions } from "#/components/form/DocumentTypeOptions";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
const documentFragment = graphql`
  fragment DocumentLayoutDrawer_documentFragment on Document {
    id
    status
    archivedAt
    canUpdate: permission(action: "core:document:update")
  }
`;

const versionFragment = graphql`
  fragment DocumentLayoutDrawer_versionFragment on DocumentVersion {
    id
    documentType
    classification
    major
    minor
    status
    updatedAt
    publishedAt
  }
`;

const updateDocumentTypeMutation = graphql`
  mutation DocumentLayoutDrawerMutation($input: UpdateDocumentVersionInput!) {
    updateDocumentVersion(input: $input) {
      documentVersion {
        id
        documentType
      }
    }
  }
`;

const updateClassificationMutation = graphql`
  mutation DocumentLayoutDrawer_updateClassificationMutation($input: UpdateDocumentVersionInput!) {
    updateDocumentVersion(input: $input) {
      documentVersion {
        id
        classification
      }
    }
  }
`;

const schema = z.object({
  documentType: z.enum(documentTypes),
});

const classificationSchema = z.object({
  classification: z.enum(documentClassifications),
});

export function DocumentLayoutDrawer(props: {
  documentFragmentRef: DocumentLayoutDrawer_documentFragment$key;
  versionFragmentRef: DocumentLayoutDrawer_versionFragment$key;
}) {
  const { documentFragmentRef, versionFragmentRef } = props;

  const { __ } = useTranslate();

  const [isEditingType, setIsEditingType] = useState(false);
  const [isEditingClassification, setIsEditingClassification] = useState(false);

  const { toast } = useToast();
  const document = useFragment<DocumentLayoutDrawer_documentFragment$key>(documentFragment, documentFragmentRef);
  const version = useFragment<DocumentLayoutDrawer_versionFragment$key>(versionFragment, versionFragmentRef);

  const isDraft = version.status === "DRAFT";
  const canEdit = document.canUpdate;

  const { control, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        documentType: version.documentType,
      },
    },
  );

  const {
    control: classificationControl,
    handleSubmit: handleClassificationSubmit,
    reset: resetClassification,
  } = useFormWithSchema(
    classificationSchema,
    {
      defaultValues: {
        classification: version.classification,
      },
    },
  );

  const [updateDocumentType, isUpdatingDocumentType]
    = useMutation<DocumentLayoutDrawerMutation>(updateDocumentTypeMutation);

  const [updateClassification, isUpdatingClassification]
    = useMutation<DocumentLayoutDrawer_updateClassificationMutation>(updateClassificationMutation);

  const handleUpdateDocumentType = (data: {
    documentType: (typeof documentTypes)[number];
  }) => {
    updateDocumentType({
      variables: {
        input: {
          documentVersionId: version.id,
          documentType: data.documentType,
        },
      },
      onCompleted: () => {
        setIsEditingType(false);
        toast({
          title: __("Success"),
          description: __("Document type updated successfully"),
          variant: "success",
        });
      },
      onError: () => {
        toast({
          title: __("Error"),
          description: __("Failed to update document type"),
          variant: "error",
        });
      },
    });
  };

  const handleUpdateClassification = (data: {
    classification: (typeof documentClassifications)[number];
  }) => {
    updateClassification({
      variables: {
        input: {
          documentVersionId: version.id,
          classification: data.classification,
        },
      },
      onCompleted: () => {
        setIsEditingClassification(false);
        toast({
          title: __("Success"),
          description: __("Document classification updated successfully"),
          variant: "success",
        });
      },
      onError: () => {
        toast({
          title: __("Error"),
          description: __("Failed to update document classification"),
          variant: "error",
        });
      },
    });
  };

  return (
    <Drawer>
      <div className="text-base text-txt-primary font-medium mb-4">
        {__("Properties")}
      </div>
      <PropertyRow label={__("Type")}>
        {isEditingType
          ? (
              <EditablePropertyContent
                onSave={() => void handleSubmit(handleUpdateDocumentType)()}
                onCancel={() => {
                  setIsEditingType(false);
                  reset();
                }}
                disabled={isUpdatingDocumentType}
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
                canEdit={canEdit && isDraft}
              >
                <div className="text-sm text-txt-secondary">
                  {getDocumentTypeLabel(__, version.documentType)}
                </div>
              </ReadOnlyPropertyContent>
            )}
      </PropertyRow>
      <PropertyRow label={__("Classification")}>
        {isEditingClassification
          ? (
              <EditablePropertyContent
                onSave={() => void handleClassificationSubmit(handleUpdateClassification)()}
                onCancel={() => {
                  setIsEditingClassification(false);
                  resetClassification();
                }}
                disabled={isUpdatingClassification}
              >
                <ControlledField
                  name="classification"
                  control={classificationControl}
                  type="select"
                >
                  <DocumentClassificationOptions />
                </ControlledField>
              </EditablePropertyContent>
            )
          : (
              <ReadOnlyPropertyContent
                onEdit={() => setIsEditingClassification(true)}
                canEdit={canEdit && isDraft}
              >
                <div className="text-sm text-txt-secondary">
                  {getDocumentClassificationLabel(__, version.classification)}
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
          {version.major}
          .
          {version.minor}
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
      {document.archivedAt && (
        <PropertyRow label={__("Archived on")}>
          <Badge variant="danger" size="md" className="gap-2">
            {formatDate(document.archivedAt)}
          </Badge>
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
