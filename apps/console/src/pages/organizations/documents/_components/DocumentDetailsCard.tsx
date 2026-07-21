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

import { documentClassifications, documentTypes } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import { Badge, Button, Card, IconCheckmark1, IconCrossLargeX, IconPencil, useToast } from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { DocumentDetailsCard_documentFragment$key } from "#/__generated__/core/DocumentDetailsCard_documentFragment.graphql";
import type { DocumentDetailsCard_updateApproversMutation } from "#/__generated__/core/DocumentDetailsCard_updateApproversMutation.graphql";
import type { DocumentDetailsCard_updateClassificationMutation } from "#/__generated__/core/DocumentDetailsCard_updateClassificationMutation.graphql";
import type { DocumentDetailsCard_versionFragment$key } from "#/__generated__/core/DocumentDetailsCard_versionFragment.graphql";
import type { DocumentDetailsCardMutation } from "#/__generated__/core/DocumentDetailsCardMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { DocumentClassificationOptions } from "#/components/form/DocumentClassificationOptions";
import { DocumentTypeOptions } from "#/components/form/DocumentTypeOptions";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const documentFragment = graphql`
  fragment DocumentDetailsCard_documentFragment on Document {
    id
    archivedAt
    canUpdate: permission(action: "core:document:update")
    defaultApprovers {
      id
      fullName
      emailAddress
    }
  }
`;

const versionFragment = graphql`
  fragment DocumentDetailsCard_versionFragment on DocumentVersion {
    id
    documentType
    classification
    major
    minor
    updatedAt
    publishedAt
  }
`;

const updateDocumentTypeMutation = graphql`
  mutation DocumentDetailsCardMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
        versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              documentType
            }
          }
        }
      }
    }
  }
`;

const updateClassificationMutation = graphql`
  mutation DocumentDetailsCard_updateClassificationMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
        versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              classification
            }
          }
        }
      }
    }
  }
`;

const updateApproversMutation = graphql`
  mutation DocumentDetailsCard_updateApproversMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
        defaultApprovers {
          id
          fullName
          emailAddress
        }
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

const approversSchema = z.object({
  approverIds: z.array(z.string()),
});

export function DocumentDetailsCard(props: {
  documentFragmentRef: DocumentDetailsCard_documentFragment$key;
  versionFragmentRef: DocumentDetailsCard_versionFragment$key;
  isEditable: boolean;
  isLatestVersion?: boolean;
  onDocumentUpdated: () => void;
}) {
  const {
    documentFragmentRef,
    versionFragmentRef,
    isEditable,
    isLatestVersion = true,
    onDocumentUpdated,
  } = props;

  const { t, i18n } = useTranslation();
  const organizationId = useOrganizationId();

  const [isEditingType, setIsEditingType] = useState(false);
  const [isEditingClassification, setIsEditingClassification] = useState(false);
  const [isEditingApprovers, setIsEditingApprovers] = useState(false);

  const { toast } = useToast();
  const document = useFragment<DocumentDetailsCard_documentFragment$key>(documentFragment, documentFragmentRef);
  const version = useFragment<DocumentDetailsCard_versionFragment$key>(versionFragment, versionFragmentRef);

  const canEdit = document.canUpdate && isEditable;
  const canEditVersionFields = canEdit;
  const canEditApprovers = document.canUpdate && isLatestVersion;

  const { control, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      values: {
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
      values: {
        classification: version.classification,
      },
    },
  );

  const {
    control: approversControl,
    handleSubmit: handleApproversSubmit,
    reset: resetApprovers,
  } = useFormWithSchema(
    approversSchema,
    {
      values: {
        approverIds: document.defaultApprovers.map(a => a.id),
      },
    },
  );

  const [updateDocumentType, isUpdatingDocumentType]
    = useMutation<DocumentDetailsCardMutation>(updateDocumentTypeMutation);

  const [updateClassification, isUpdatingClassification]
    = useMutation<DocumentDetailsCard_updateClassificationMutation>(updateClassificationMutation);

  const [updateApprovers, isUpdatingApprovers]
    = useMutation<DocumentDetailsCard_updateApproversMutation>(updateApproversMutation);

  const handleUpdateDocumentType = (data: {
    documentType: (typeof documentTypes)[number];
  }) => {
    updateDocumentType({
      variables: {
        input: {
          id: document.id,
          documentType: data.documentType,
        },
      },
      onCompleted: () => {
        setIsEditingType(false);
        onDocumentUpdated();
        toast({
          title: t("documentDetails.messages.successTitle"),
          description: t("documentDetails.messages.typeUpdated"),
          variant: "success",
        });
      },
      onError: () => {
        toast({
          title: t("documentDetails.errors.title"),
          description: t("documentDetails.errors.updateType"),
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
          id: document.id,
          classification: data.classification,
        },
      },
      onCompleted: () => {
        setIsEditingClassification(false);
        onDocumentUpdated();
        toast({
          title: t("documentDetails.messages.successTitle"),
          description: t("documentDetails.messages.classificationUpdated"),
          variant: "success",
        });
      },
      onError: () => {
        toast({
          title: t("documentDetails.errors.title"),
          description: t("documentDetails.errors.updateClassification"),
          variant: "error",
        });
      },
    });
  };

  const handleUpdateApprovers = (data: { approverIds: string[] }) => {
    updateApprovers({
      variables: {
        input: {
          id: document.id,
          defaultApproverIds: data.approverIds,
        },
      },
      onCompleted: () => {
        setIsEditingApprovers(false);
        toast({
          title: t("documentDetails.messages.successTitle"),
          description: t("documentDetails.messages.approversUpdated"),
          variant: "success",
        });
      },
      onError: () => {
        toast({
          title: t("documentDetails.errors.title"),
          description: t("documentDetails.errors.updateApprovers"),
          variant: "error",
        });
      },
    });
  };

  return (
    <Card className="space-y-4" padded>
      <div className="grid grid-cols-3 gap-4">
        <div>
          <div className="text-xs text-txt-tertiary font-semibold mb-1">
            {t("documentDetails.fields.type")}
          </div>
          {isEditingType
            ? (
                <div className="flex items-center gap-2">
                  <div className="flex-1">
                    <ControlledField
                      name="documentType"
                      control={control}
                      type="select"
                    >
                      <DocumentTypeOptions />
                    </ControlledField>
                  </div>
                  <Button
                    variant="quaternary"
                    icon={IconCheckmark1}
                    onClick={() => void handleSubmit(handleUpdateDocumentType)()}
                    disabled={isUpdatingDocumentType}
                  />
                  <Button
                    variant="quaternary"
                    icon={IconCrossLargeX}
                    onClick={() => {
                      setIsEditingType(false);
                      reset();
                    }}
                  />
                </div>
              )
            : (
                <div className="flex items-center gap-2">
                  <div className="text-sm text-txt-primary">
                    {t(`documentDetails.documentTypes.${version.documentType.toLowerCase()}`)}
                  </div>
                  {canEditVersionFields && (
                    <Button
                      variant="quaternary"
                      icon={IconPencil}
                      onClick={() => setIsEditingType(true)}
                    />
                  )}
                </div>
              )}
        </div>
        <div>
          <div className="text-xs text-txt-tertiary font-semibold mb-1">
            {t("documentDetails.fields.classification")}
          </div>
          {isEditingClassification
            ? (
                <div className="flex items-center gap-2">
                  <div className="flex-1">
                    <ControlledField
                      name="classification"
                      control={classificationControl}
                      type="select"
                    >
                      <DocumentClassificationOptions />
                    </ControlledField>
                  </div>
                  <Button
                    variant="quaternary"
                    icon={IconCheckmark1}
                    onClick={() => void handleClassificationSubmit(handleUpdateClassification)()}
                    disabled={isUpdatingClassification}
                  />
                  <Button
                    variant="quaternary"
                    icon={IconCrossLargeX}
                    onClick={() => {
                      setIsEditingClassification(false);
                      resetClassification();
                    }}
                  />
                </div>
              )
            : (
                <div className="flex items-center gap-2">
                  <div className="text-sm text-txt-primary">
                    {t(`documentDetails.classifications.${version.classification.toLowerCase()}`)}
                  </div>
                  {canEditVersionFields && (
                    <Button
                      variant="quaternary"
                      icon={IconPencil}
                      onClick={() => setIsEditingClassification(true)}
                    />
                  )}
                </div>
              )}
        </div>
        {isLatestVersion && (
          <div>
            <div className="text-xs text-txt-tertiary font-semibold mb-1">
              {t("documentDetails.fields.approvers")}
            </div>
            {isEditingApprovers
              ? (
                  <div className="flex items-center gap-2">
                    <div className="flex-1">
                      <PeopleMultiSelectField
                        name="approverIds"
                        control={approversControl}
                        organizationId={organizationId}
                        selectedPeople={document.defaultApprovers.map(a => ({
                          id: a.id,
                          fullName: a.fullName,
                          emailAddress: a.emailAddress,
                        }))}
                        placeholder={t("documentDetails.fields.approversPlaceholder")}
                      />
                    </div>
                    <Button
                      variant="quaternary"
                      icon={IconCheckmark1}
                      onClick={() => void handleApproversSubmit(handleUpdateApprovers)()}
                      disabled={isUpdatingApprovers}
                    />
                    <Button
                      variant="quaternary"
                      icon={IconCrossLargeX}
                      onClick={() => {
                        setIsEditingApprovers(false);
                        resetApprovers({ approverIds: document.defaultApprovers.map(a => a.id) });
                      }}
                    />
                  </div>
                )
              : (
                  <div className="flex items-center gap-2">
                    <div className="text-sm text-txt-primary">
                      {document.defaultApprovers.length > 0
                        ? document.defaultApprovers.map(a => a.fullName).join(", ")
                        : t("documentDetails.none")}
                    </div>
                    {canEditApprovers && (
                      <Button
                        variant="quaternary"
                        icon={IconPencil}
                        onClick={() => setIsEditingApprovers(true)}
                      />
                    )}
                  </div>
                )}
          </div>
        )}
      </div>
      <div className="grid grid-cols-3 gap-4">
        <div>
          <div className="text-xs text-txt-tertiary font-semibold mb-1">
            {t("documentDetails.fields.version")}
          </div>
          <div className="text-sm text-txt-primary">
            {version.major}
            .
            {version.minor}
          </div>
        </div>
        <div>
          <div className="text-xs text-txt-tertiary font-semibold mb-1">
            {t("documentDetails.fields.lastModified")}
          </div>
          <div className="text-sm text-txt-primary">
            {dateFormat(i18n.language, version.updatedAt)}
          </div>
        </div>
        <div>
          {version.publishedAt && (
            <>
              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                {t("documentDetails.fields.publishedDate")}
              </div>
              <div className="text-sm text-txt-primary">
                {dateFormat(i18n.language, version.publishedAt)}
              </div>
            </>
          )}
          {document.archivedAt && (
            <>
              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                {t("documentDetails.fields.archivedOn")}
              </div>
              <Badge variant="danger" size="md" className="gap-2">
                {dateFormat(i18n.language, document.archivedAt)}
              </Badge>
            </>
          )}
        </div>
      </div>
    </Card>
  );
}
