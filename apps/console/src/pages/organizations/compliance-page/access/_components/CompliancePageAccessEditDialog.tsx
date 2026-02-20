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
  Field,
  Spinner,
} from "@probo/ui";
import { Suspense, useCallback, useEffect, useState } from "react";
import {
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageAccessEditDialogQuery as CompliancePageAccessEditDialogQueryType } from "#/__generated__/core/CompliancePageAccessEditDialogQuery.graphql";
import type { CompliancePageAccessEditDialogUpdateMutation } from "#/__generated__/core/CompliancePageAccessEditDialogUpdateMutation.graphql";
import type { CompliancePageAccessListItemFragment$data } from "#/__generated__/core/CompliancePageAccessListItemFragment.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
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
        email
        name
        state
        hasAcceptedNonDisclosureAgreement
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
    <Dialog defaultOpen={true} title={__("Edit Access")} onClose={onClose}>
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

  const editSchema = z.object({
    name: z
      .string()
      .min(1, __("Name is required"))
      .min(2, __("Name must be at least 2 characters long")),
    active: z.boolean(),
  });
  const editForm = useFormWithSchema(editSchema, {
    defaultValues: { name: access.name, active: access.state === "ACTIVE" },
  });

  const [updateTrustCenterAccess, isUpdating] = useMutationWithToasts<CompliancePageAccessEditDialogUpdateMutation>(
    updateAccessMutation,
    {
      successMessage: __("Access updated successfully"),
      errorMessage: __("Failed to update access"),
    },
  );

  const handleSubmit = editForm.handleSubmit(async (data) => {
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
          name: data.name.trim(),
          state: data.active ? "ACTIVE" : "INACTIVE",
          documents,
          reports,
          trustCenterFiles,
        },
      },
      onSuccess: onSubmit,
    });
  });

  return (
    <form onSubmit={e => void handleSubmit(e)}>
      <DialogContent padded className="space-y-6">
        <div>
          <p className="text-txt-secondary text-sm mb-4">
            {__("Update access settings and document permissions")}
          </p>

          <Field
            label={__("Full Name")}
            required
            error={editForm.formState.errors.name?.message}
            {...editForm.register("name")}
            placeholder={__("John Doe")}
          />

        </div>

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
        <Button type="submit" disabled={isUpdating}>
          {isUpdating && <Spinner />}
          {__("Update Access")}
        </Button>
      </DialogFooter>
    </form>
  );
}
