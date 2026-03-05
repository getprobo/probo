import type { TrustCenterDocumentAccessStatus } from "@probo/coredata";
import {
  getTrustCenterDocumentAccessInfo,
  type TrustCenterDocumentAccessInfo,
} from "@probo/helpers";
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
import { graphql } from "relay-runtime";

import type { CompliancePageAccessEditDialogQuery as CompliancePageAccessEditDialogQueryType } from "#/__generated__/core/CompliancePageAccessEditDialogQuery.graphql";
import type { CompliancePageAccessEditDialogUpdateMutation } from "#/__generated__/core/CompliancePageAccessEditDialogUpdateMutation.graphql";
import type { CompliancePageAccessListItemFragment$data } from "#/__generated__/core/CompliancePageAccessListItemFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { CompliancePageDocumentAccessList } from "#/pages/organizations/compliance-page/access/_components/CompliancePageDocumentAccessList";
import { ElectronicSignatureSection } from "#/pages/organizations/compliance-page/access/_components/ElectronicSignatureSection";

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
              id
              status
              document {
                id
                title
                documentType
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
