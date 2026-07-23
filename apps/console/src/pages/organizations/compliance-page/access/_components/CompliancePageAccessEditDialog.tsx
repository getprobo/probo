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

import type { CompliancePortalDocumentAccessStatus } from "@probo/coredata";
import type { CompliancePageDocumentAccessInfo } from "@probo/helpers";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
} from "@probo/ui";
import { Suspense, useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { graphql, readInlineData } from "relay-runtime";

import type { CompliancePageAccessEditDialogDocumentAccessFragment$data, CompliancePageAccessEditDialogDocumentAccessFragment$key } from "#/__generated__/core/CompliancePageAccessEditDialogDocumentAccessFragment.graphql";
import type { CompliancePageAccessEditDialogQuery as CompliancePageAccessEditDialogQueryType } from "#/__generated__/core/CompliancePageAccessEditDialogQuery.graphql";
import type { CompliancePageAccessEditDialogUpdateMutation } from "#/__generated__/core/CompliancePageAccessEditDialogUpdateMutation.graphql";
import type { CompliancePageAccessListItemFragment$data } from "#/__generated__/core/CompliancePageAccessListItemFragment.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { CompliancePageDocumentAccessList } from "#/pages/organizations/compliance-page/access/_components/CompliancePageDocumentAccessList";
import { ElectronicSignatureSection } from "#/pages/organizations/compliance-page/access/_components/ElectronicSignatureSection";

const documentAccessFragment = graphql`
  fragment CompliancePageAccessEditDialogDocumentAccessFragment on CompliancePortalDocumentAccess @inline {
    id
    status
    document {
      id
      versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
        edges {
          node {
            title
            documentType
          }
        }
      }
    }
    reportFile {
      id
      fileName
    }
    audit {
      framework {
        name
      }
    }
    compliancePortalFile {
      id
      name
      category
    }
  }
`;

function getCompliancePageDocumentAccessInfo(
  fragmentRef: CompliancePageAccessEditDialogDocumentAccessFragment$key,
  t: (key: string) => string,
): CompliancePageDocumentAccessInfo {
  const node = readInlineData(documentAccessFragment, fragmentRef);
  return toDocumentAccessInfo(node, t);
}

function toDocumentAccessInfo(
  node: CompliancePageAccessEditDialogDocumentAccessFragment$data,
  t: (key: string) => string,
): CompliancePageDocumentAccessInfo {
  if (node.document) {
    return {
      persisted: node.id !== node.document.id,
      variant: "info",
      name: node.document.versions?.edges[0]?.node.title ?? "",
      type: "document",
      typeLabel: t("accessEditDialog.types.document"),
      category: node.document.versions?.edges[0]?.node.documentType ?? "",
      id: node.document.id,
      status: node.status,
    };
  }
  if (node.reportFile) {
    return {
      persisted: node.id !== node.reportFile.id,
      variant: "success",
      name: node.reportFile.fileName,
      type: "report",
      typeLabel: t("accessEditDialog.types.report"),
      category: node.audit?.framework?.name ?? "",
      id: node.reportFile.id,
      status: node.status,
    };
  }
  if (node.compliancePortalFile) {
    return {
      persisted: node.id !== node.compliancePortalFile.id,
      variant: "highlight",
      name: node.compliancePortalFile.name,
      type: "file",
      typeLabel: t("accessEditDialog.types.file"),
      category: node.compliancePortalFile.category,
      id: node.compliancePortalFile.id,
      status: node.status,
    };
  }
  throw new Error("Unknown compliance page access document type");
}

const compliancePageAccessEditDialogQuery = graphql`
  query CompliancePageAccessEditDialogQuery($accessId: ID!) {
    node(id: $accessId) {
      ... on CompliancePortalAccess {
        id
        ndaSignature {
          ...ElectronicSignatureSectionFragment
        }
        availableDocumentAccesses(
          first: 100
          orderBy: { field: CREATED_AT, direction: DESC }
        ) {
          edges {
            node {
              ...CompliancePageAccessEditDialogDocumentAccessFragment
            }
          }
        }
      }
    }
  }
`;

const updateAccessMutation = graphql`
  mutation CompliancePageAccessEditDialogUpdateMutation(
    $input: UpdateCompliancePortalAccessInput!
  ) {
    updateCompliancePortalAccess(input: $input) {
      compliancePortalAccess {
        id
        createdAt
        updatedAt
        pendingRequestCount
        activeCount
      }
    }
  }
`;

export function CompliancePageAccessEditDialog(props: {
  access: CompliancePageAccessListItemFragment$data;
  onClose: () => void;
}) {
  const { access, onClose } = props;

  const { t } = useTranslation("organizations/compliance-page");

  const [queryRef, loadQuery]
    = useQueryLoader<CompliancePageAccessEditDialogQueryType>(
      compliancePageAccessEditDialogQuery,
    );

  useEffect(() => {
    loadQuery(
      {
        accessId: access.id,
      },
      {
        fetchPolicy: "network-only",
      },
    );
  }, [access.id, loadQuery]);

  return (
    <Dialog
      defaultOpen={true}
      title={t("accessEditDialog.title", { email: access.profile.emailAddress })}
      onClose={onClose}
    >
      {queryRef && (
        <Suspense>
          <CompliancePageAccessEditForm
            access={access}
            queryRef={queryRef}
            onSubmit={onClose}
          />
        </Suspense>
      )}
    </Dialog>
  );
}

function CompliancePageAccessEditForm(props: {
  access: CompliancePageAccessListItemFragment$data;
  onSubmit: () => void;
  queryRef: PreloadedQuery<CompliancePageAccessEditDialogQueryType>;
}) {
  const { access, onSubmit, queryRef } = props;

  const { t } = useTranslation("organizations/compliance-page");
  const data
    = usePreloadedQuery<CompliancePageAccessEditDialogQueryType>(
      compliancePageAccessEditDialogQuery,
      queryRef,
    );

  const initialDocumentAccesses
    = data.node.availableDocumentAccesses?.edges.map(edge =>
      getCompliancePageDocumentAccessInfo(edge.node, t),
    ) ?? [];
  const initialStatusByID = initialDocumentAccesses.reduce<
    Record<string, CompliancePortalDocumentAccessStatus>
  >((acc, docAccess) => {
    acc[docAccess.id] = docAccess.status;
    return acc;
  }, {});
  const [documentAccesses, setDocumentAccesses] = useState<
    CompliancePageDocumentAccessInfo[]
  >(initialDocumentAccesses);

  const handleUpdateDocumentAccessStatus = useCallback(
    (
      documentAccess: CompliancePageDocumentAccessInfo,
      status: CompliancePortalDocumentAccessStatus,
    ) => {
      setDocumentAccesses((prev) => {
        const nextDocumentAccesses = [...prev];
        const docAccessIndex = nextDocumentAccesses.findIndex(
          element => element.id === documentAccess.id,
        );
        const previousDocAccess = nextDocumentAccesses[docAccessIndex];
        nextDocumentAccesses.splice(docAccessIndex, 1, {
          ...previousDocAccess,
          status,
        });

        return nextDocumentAccesses;
      });
    },
    [],
  );
  const handleGrantAllDocumentAccess = useCallback(() => {
    setDocumentAccesses(prev =>
      prev.map(element => ({ ...element, status: "GRANTED" })),
    );
  }, []);
  const handleRejectOrRevokeAllDocumentAccess = useCallback(() => {
    setDocumentAccesses(prev =>
      prev.map(element => ({
        ...element,
        status:
          initialStatusByID[element.id] === "GRANTED" ? "REVOKED" : "REJECTED",
      })),
    );
  }, [initialStatusByID]);

  const [updateCompliancePageAccess, isUpdating] = useMutation<CompliancePageAccessEditDialogUpdateMutation>(
    updateAccessMutation,
    {
      successMessage: t("accessEditDialog.messages.updated"),
      errorToast: t("accessEditDialog.errors.update"),
    },
  );

  const handleSubmit = async () => {
    const documents: { id: string; status: CompliancePortalDocumentAccessStatus }[]
      = [];
    const reports: { id: string; status: CompliancePortalDocumentAccessStatus }[]
      = [];
    const compliancePageFiles: {
      id: string;
      status: CompliancePortalDocumentAccessStatus;
    }[] = [];

    for (const docAccess of documentAccesses) {
      if (docAccess.persisted || docAccess.status !== "REQUESTED") {
        switch (docAccess.type) {
          case "document":
            documents.push({ id: docAccess.id, status: docAccess.status });
            break;
          case "report":
            reports.push({ id: docAccess.id, status: docAccess.status });
            break;
          case "file":
            compliancePageFiles.push({
              id: docAccess.id,
              status: docAccess.status,
            });
            break;
        }
      }
    }

    await updateCompliancePageAccess({
      variables: {
        input: {
          id: access.id,
          documents,
          reports,
          compliancePortalFiles: compliancePageFiles,
        },
      },
    });

    onSubmit();
  };

  return (
    <>
      <DialogContent padded className="space-y-6">
        {data.node.ndaSignature && (
          <ElectronicSignatureSection fragmentRef={data.node.ndaSignature} />
        )}

        <CompliancePageDocumentAccessList
          documentAccesses={documentAccesses}
          initialStatusByID={initialStatusByID}
          onGrantAll={handleGrantAllDocumentAccess}
          onRejectOrRevokeAll={handleRejectOrRevokeAllDocumentAccess}
          onUpdateStatus={handleUpdateDocumentAccessStatus}
        />
      </DialogContent>

      <DialogFooter>
        <Button type="button" disabled={isUpdating} onClick={() => void handleSubmit()}>
          {isUpdating && <Spinner />}
          {t("accessEditDialog.actions.update")}
        </Button>
      </DialogFooter>
    </>
  );
}
