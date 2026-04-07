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

import type { TrustCenterDocumentAccessStatus } from "@probo/coredata";
import type { TrustCenterDocumentAccessInfo } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
} from "@probo/ui";
import { Suspense, useCallback, useEffect, useState } from "react";
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
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { CompliancePageDocumentAccessList } from "#/pages/organizations/compliance-page/access/_components/CompliancePageDocumentAccessList";
import { ElectronicSignatureSection } from "#/pages/organizations/compliance-page/access/_components/ElectronicSignatureSection";

const documentAccessFragment = graphql`
  fragment CompliancePageAccessEditDialogDocumentAccessFragment on TrustCenterDocumentAccess @inline {
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
    report {
      id
      filename
      audit {
        id
        framework {
          name
        }
      }
    }
    trustCenterFile {
      id
      name
      category
    }
  }
`;

function getTrustCenterDocumentAccessInfo(
  fragmentRef: CompliancePageAccessEditDialogDocumentAccessFragment$key,
  __: (key: string) => string,
): TrustCenterDocumentAccessInfo {
  const node = readInlineData(documentAccessFragment, fragmentRef);
  return toDocumentAccessInfo(node, __);
}

function toDocumentAccessInfo(
  node: CompliancePageAccessEditDialogDocumentAccessFragment$data,
  __: (key: string) => string,
): TrustCenterDocumentAccessInfo {
  if (node.document) {
    return {
      persisted: node.id !== node.document.id,
      variant: "info",
      name: node.document.versions?.edges[0]?.node.title ?? "",
      type: "document",
      typeLabel: __("Document"),
      category: node.document.versions?.edges[0]?.node.documentType ?? "",
      id: node.document.id,
      status: node.status,
    };
  }
  if (node.report) {
    return {
      persisted: node.id !== node.report.id,
      variant: "success",
      name: node.report.filename,
      type: "report",
      typeLabel: __("Report"),
      category: node.report.audit?.framework?.name ?? "",
      id: node.report.id,
      status: node.status,
    };
  }
  if (node.trustCenterFile) {
    return {
      persisted: node.id !== node.trustCenterFile.id,
      variant: "highlight",
      name: node.trustCenterFile.name,
      type: "file",
      typeLabel: __("File"),
      category: node.trustCenterFile.category,
      id: node.trustCenterFile.id,
      status: node.status,
    };
  }
  throw new Error("Unknown trust center access document type");
}

const compliancePageAccessEditDialogQuery = graphql`
  query CompliancePageAccessEditDialogQuery($accessId: ID!) {
    node(id: $accessId) {
      ... on TrustCenterAccess {
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
    $input: UpdateTrustCenterAccessInput!
  ) {
    updateTrustCenterAccess(input: $input) {
      trustCenterAccess {
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

  const { __ } = useTranslate();

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
    <Dialog defaultOpen={true} title={__(`Edit Access for ${access.profile.emailAddress}`)} onClose={onClose}>
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

  const { __ } = useTranslate();
  const data
    = usePreloadedQuery<CompliancePageAccessEditDialogQueryType>(
      compliancePageAccessEditDialogQuery,
      queryRef,
    );

  const initialDocumentAccesses
    = data.node.availableDocumentAccesses?.edges.map(edge =>
      getTrustCenterDocumentAccessInfo(edge.node, __),
    ) ?? [];
  const initialStatusByID = initialDocumentAccesses.reduce<
    Record<string, TrustCenterDocumentAccessStatus>
  >((acc, docAccess) => {
    acc[docAccess.id] = docAccess.status;
    return acc;
  }, {});
  const [documentAccesses, setDocumentAccesses] = useState<
    TrustCenterDocumentAccessInfo[]
  >(initialDocumentAccesses);

  const handleUpdateDocumentAccessStatus = useCallback(
    (
      documentAccess: TrustCenterDocumentAccessInfo,
      status: TrustCenterDocumentAccessStatus,
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

  const [updateTrustCenterAccess, isUpdating] = useMutationWithToasts<CompliancePageAccessEditDialogUpdateMutation>(
    updateAccessMutation,
    {
      successMessage: __("Access updated successfully"),
      errorMessage: __("Failed to update access"),
    },
  );

  const handleSubmit = async () => {
    const documents: { id: string; status: TrustCenterDocumentAccessStatus }[]
      = [];
    const reports: { id: string; status: TrustCenterDocumentAccessStatus }[]
      = [];
    const trustCenterFiles: {
      id: string;
      status: TrustCenterDocumentAccessStatus;
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
            trustCenterFiles.push({
              id: docAccess.id,
              status: docAccess.status,
            });
            break;
        }
      }
    }

    await updateTrustCenterAccess({
      variables: {
        input: {
          id: access.id,
          documents,
          reports,
          trustCenterFiles,
        },
      },
      onSuccess: onSubmit,
    });
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
          {__("Update Access")}
        </Button>
      </DialogFooter>
    </>
  );
}
