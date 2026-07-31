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

import { getCompliancePortalLinkedVisibilityOptions } from "@probo/helpers";
import { Badge, Button, Dialog, DialogContent, DialogFooter, Field, Option, useDialogRef } from "@probo/ui";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { type DataID, graphql, useLazyLoadQuery } from "react-relay";

import type { NewCompliancePortalDocumentDialog_addMutation } from "#/__generated__/core/NewCompliancePortalDocumentDialog_addMutation.graphql";
import type { NewCompliancePortalDocumentDialogQuery } from "#/__generated__/core/NewCompliancePortalDocumentDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const pickerQuery = graphql`
  query NewCompliancePortalDocumentDialogQuery($organizationId: ID!, $compliancePortalId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        documents(first: 100, orderBy: { field: UPDATED_AT, direction: DESC }) {
          edges {
            node {
              id
              currentPublishedMajor
              latestPublishedVersion: versions(
                first: 1
                orderBy: { field: CREATED_AT, direction: DESC }
                filter: { statuses: [PUBLISHED] }
              ) {
                edges {
                  node {
                    title
                  }
                }
              }
            }
          }
        }
      }
    }
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        documents(first: 100) {
          edges {
            node {
              document {
                id
              }
            }
          }
        }
      }
    }
  }
`;

const addMutation = graphql`
  mutation NewCompliancePortalDocumentDialog_addMutation(
    $input: UpdateCompliancePortalDocumentVisibilityInput!
    $connections: [ID!]!
  ) {
    updateCompliancePortalDocumentVisibility(input: $input) {
      catalogDocumentEdge @appendEdge(connections: $connections) {
        cursor
        node {
          ...CompliancePortalDocumentListItem_catalogDocumentFragment
        }
      }
    }
  }
`;

function DocumentPicker(props: {
  compliancePortalId: string;
  connectionId: DataID;
  onClose: () => void;
}) {
  const { compliancePortalId, connectionId, onClose } = props;
  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");
  const visibilityOptions = getCompliancePortalLinkedVisibilityOptions(t);

  const data = useLazyLoadQuery<NewCompliancePortalDocumentDialogQuery>(pickerQuery, {
    organizationId,
    compliancePortalId,
  });

  const [documentId, setDocumentId] = useState("");
  const [visibility, setVisibility] = useState<"RESTRICTED" | "PUBLIC">("RESTRICTED");

  const [addDocument, isAdding] = useMutation<NewCompliancePortalDocumentDialog_addMutation>(
    addMutation,
    {
      successMessage: t("documentList.addDialog.success"),
      errorToast: t("documentList.addDialog.error"),
    },
  );

  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid organization node");
  }

  if (data.compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid compliance portal node");
  }

  const linkedDocumentIDs = new Set(
    data.compliancePortal.documents.edges.map(({ node }) => node.document.id),
  );

  const publishedDocuments = data.organization.documents.edges.filter(
    ({ node }) =>
      node.currentPublishedMajor != null
      && node.latestPublishedVersion.edges.length > 0
      && !linkedDocumentIDs.has(node.id),
  );

  const handleSubmit = async () => {
    if (documentId === "") {
      return;
    }

    await addDocument({
      variables: {
        connections: [connectionId],
        input: {
          compliancePortalId,
          documentId,
          compliancePortalVisibility: visibility,
        },
      },
    });
    onClose();
  };

  return (
    <>
      <DialogContent className="space-y-4">
        <Field
          type="select"
          label={t("documentList.addDialog.documentLabel")}
          value={documentId}
          onValueChange={value => setDocumentId(typeof value === "string" ? value : "")}
        >
          {publishedDocuments.map(({ node }) => (
            <Option key={node.id} value={node.id}>
              {node.latestPublishedVersion.edges[0]?.node.title ?? node.id}
            </Option>
          ))}
        </Field>
        <Field
          type="select"
          label={t("documentList.addDialog.visibilityLabel")}
          value={visibility}
          onValueChange={value => setVisibility(value === "PUBLIC" ? "PUBLIC" : "RESTRICTED")}
        >
          {visibilityOptions.map(option => (
            <Option key={option.value} value={option.value}>
              <Badge variant={option.variant}>{option.label}</Badge>
            </Option>
          ))}
        </Field>
      </DialogContent>
      <DialogFooter>
        <Button variant="secondary" onClick={onClose}>
          {t("documentList.addDialog.cancel")}
        </Button>
        <Button
          disabled={isAdding || documentId === ""}
          onClick={() => void handleSubmit()}
        >
          {t("documentList.addDialog.submit")}
        </Button>
      </DialogFooter>
    </>
  );
}

export function NewCompliancePortalDocumentDialog(props: {
  compliancePortalId: string;
  connectionId: DataID;
  disabled?: boolean;
}) {
  const { compliancePortalId, connectionId, disabled } = props;
  const { t } = useTranslation("organizations/compliance-portals");
  const dialogRef = useDialogRef();
  const [open, setOpen] = useState(false);

  return (
    <Dialog
      ref={dialogRef}
      trigger={(
        <Button variant="secondary" disabled={disabled}>
          {t("documentList.addDocument")}
        </Button>
      )}
      onOpenChange={setOpen}
      title={t("documentList.addDialog.title")}
    >
      {open && (
        <Suspense fallback={<p className="text-sm text-txt-secondary">{t("documentList.addDialog.loading")}</p>}>
          <DocumentPicker
            compliancePortalId={compliancePortalId}
            connectionId={connectionId}
            onClose={() => dialogRef.current?.close()}
          />
        </Suspense>
      )}
    </Dialog>
  );
}
