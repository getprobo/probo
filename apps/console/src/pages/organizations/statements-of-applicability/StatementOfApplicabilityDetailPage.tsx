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

import { formatDate, formatError, type GraphQLError, promisifyMutation, sprintf, validateSnapshotConsistency } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  IconArrowDown,
  IconCheckmark1,
  IconCrossLargeX,
  IconPencil,
  IconTrashCan,
  Input,
  PageHeader,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Suspense, useState } from "react";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate, useParams } from "react-router";
import { z } from "zod";

import type { StatementOfApplicabilityDetailPageDeleteMutation } from "#/__generated__/core/StatementOfApplicabilityDetailPageDeleteMutation.graphql";
import type { StatementOfApplicabilityDetailPageExportMutation } from "#/__generated__/core/StatementOfApplicabilityDetailPageExportMutation.graphql";
import type { StatementOfApplicabilityDetailPageQuery } from "#/__generated__/core/StatementOfApplicabilityDetailPageQuery.graphql";
import type { StatementOfApplicabilityDetailPageUpdateMutation } from "#/__generated__/core/StatementOfApplicabilityDetailPageUpdateMutation.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import StatementOfApplicabilityControlsTab from "./tabs/StatementOfApplicabilityControlsTab";

export const statementOfApplicabilityDetailPageQuery = graphql`
    query StatementOfApplicabilityDetailPageQuery($statementOfApplicabilityId: ID!) {
        node(id: $statementOfApplicabilityId) {
            ... on StatementOfApplicability {
                id
                name
                snapshotId
                createdAt
                updatedAt
                canUpdate: permission(action: "core:statement-of-applicability:update")
                canDelete: permission(action: "core:statement-of-applicability:delete")
                canExport: permission(action: "core:statement-of-applicability:export")
                owner {
                    id
                    fullName
                }
                ...StatementOfApplicabilityControlsTabFragment
            }
        }
    }
`;

const exportMutation = graphql`
    mutation StatementOfApplicabilityDetailPageExportMutation(
        $input: ExportStatementOfApplicabilityPDFInput!
    ) {
        exportStatementOfApplicabilityPDF(input: $input) {
            data
        }
    }
`;

const updateMutation = graphql`
    mutation StatementOfApplicabilityDetailPageUpdateMutation(
        $input: UpdateStatementOfApplicabilityInput!
    ) {
        updateStatementOfApplicability(input: $input) {
            statementOfApplicability {
                id
                name
                sourceId
                snapshotId
                createdAt
                updatedAt
                owner {
                    id
                    fullName
                }
            }
        }
    }
`;

const deleteMutation = graphql`
    mutation StatementOfApplicabilityDetailPageDeleteMutation(
        $input: DeleteStatementOfApplicabilityInput!
        $connections: [ID!]!
    ) {
        deleteStatementOfApplicability(input: $input) {
            deletedStatementOfApplicabilityId @deleteEdge(connections: $connections)
        }
    }
`;

const StatementOfApplicabilityConnectionKey = "StatementsOfApplicabilityPage_statementsOfApplicability";

type Props = {
  queryRef: PreloadedQuery<StatementOfApplicabilityDetailPageQuery>;
};

export default function StatementOfApplicabilityDetailPage(props: Props) {
  const { statementOfApplicabilityId, snapshotId } = useParams<{
    statementOfApplicabilityId: string;
    snapshotId?: string;
  }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(statementOfApplicabilityDetailPageQuery, props.queryRef);
  const statementOfApplicability = data.node;
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const isSnapshotMode = Boolean(snapshotId);
  const confirm = useConfirm();
  const { toast } = useToast();

  if (!statementOfApplicabilityId || !statementOfApplicability) {
    throw new Error(
      "Cannot load statement of applicability detail page without statementOfApplicabilityId parameter",
    );
  }

  validateSnapshotConsistency(statementOfApplicability, snapshotId);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    StatementOfApplicabilityConnectionKey,
    { filter: { snapshotId: snapshotId ?? null } },
  );

  const [deleteStatementOfApplicability]
    = useMutation<StatementOfApplicabilityDetailPageDeleteMutation>(deleteMutation);

  const handleDelete = () => {
    if (!statementOfApplicability.id || !statementOfApplicability.name) {
      return alert(__("Failed to delete statement of applicability: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(deleteStatementOfApplicability)({
          variables: {
            input: {
              statementOfApplicabilityId: statementOfApplicability.id!,
            },
            connections: [connectionId],
          },
        })
          .then(() => {
            void navigate(`/organizations/${organizationId}/statements-of-applicability`);
          })
          .catch((error) => {
            toast({
              title: __("Error"),
              description: formatError(
                __("Failed to delete statement of applicability"),
                error as GraphQLError,
              ),
              variant: "error",
            });
          }),
      {
        message: sprintf(
          __(
            "This will permanently delete \"%s\". This action cannot be undone.",
          ),
          statementOfApplicability.name,
        ),
      },
    );
  };

  usePageTitle(statementOfApplicability.name || __("Statement of Applicability"));

  const [isEditingName, setIsEditingName] = useState(false);
  const [isEditingOwner, setIsEditingOwner] = useState(false);
  const [updateStatementOfApplicability, isUpdating]
    = useMutationWithToasts<StatementOfApplicabilityDetailPageUpdateMutation>(
      updateMutation,
      {
        successMessage: __("Statement of Applicability updated successfully."),
        errorMessage: __("Failed to update Statement of Applicability"),
      },
    );

  const canUpdate = !isSnapshotMode && statementOfApplicability.canUpdate;
  const canDelete = !isSnapshotMode && statementOfApplicability.canDelete;

  const [exportStatementOfApplicabilityPDF, isExporting]
    = useMutationWithToasts<StatementOfApplicabilityDetailPageExportMutation>(
      exportMutation,
      {
        successMessage: __(
          "Statement of Applicability exported successfully.",
        ),
        errorMessage: __("Failed to export Statement of Applicability"),
      },
    );

  const handleExport = async () => {
    if (!statementOfApplicability.id) return;

    await exportStatementOfApplicabilityPDF({
      variables: {
        input: {
          statementOfApplicabilityId: statementOfApplicability.id,
        },
      },
      onCompleted: (data) => {
        if (data.exportStatementOfApplicabilityPDF?.data) {
          const link = window.document.createElement("a");
          link.href = data.exportStatementOfApplicabilityPDF.data;
          link.download = `${statementOfApplicability.name || "statement-of-applicability"}.pdf`;
          window.document.body.appendChild(link);
          link.click();
          window.document.body.removeChild(link);
        }
      },
    });
  };

  const nameSchema = z.object({
    name: z.string().min(1, __("Name is required")),
  });

  const ownerSchema = z.object({
    ownerId: z.string().min(1, __("Owner is required")),
  });

  const {
    register: registerName,
    handleSubmit: handleSubmitName,
    reset: resetName,
  } = useFormWithSchema(nameSchema, {
    defaultValues: {
      name: statementOfApplicability.name || "",
    },
  });

  const {
    control: controlOwner,
    handleSubmit: handleSubmitOwner,
    reset: resetOwner,
  } = useFormWithSchema(ownerSchema, {
    defaultValues: {
      ownerId: statementOfApplicability.owner?.id || "",
    },
  });

  const handleUpdateName = handleSubmitName(async (data) => {
    if (!statementOfApplicability.id) return;

    await updateStatementOfApplicability({
      variables: {
        input: {
          id: statementOfApplicability.id,
          name: data.name,
        },
      },
      onSuccess: () => {
        setIsEditingName(false);
        resetName({ name: data.name });
      },
    });
  });

  const handleUpdateOwner = handleSubmitOwner(async (data) => {
    if (!statementOfApplicability.id) return;

    await updateStatementOfApplicability({
      variables: {
        input: {
          id: statementOfApplicability.id,
          ownerId: data.ownerId,
        },
      },
      onSuccess: () => {
        setIsEditingOwner(false);
        resetOwner({ ownerId: data.ownerId });
      },
    });
  });

  const handleCancelNameEdit = () => {
    setIsEditingName(false);
    resetName({
      name: statementOfApplicability.name || "",
    });
  };

  const handleCancelOwnerEdit = () => {
    setIsEditingOwner(false);
    resetOwner({
      ownerId: statementOfApplicability.owner?.id || "",
    });
  };

  const listUrl = snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/statements-of-applicability`
    : `/organizations/${organizationId}/statements-of-applicability`;

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <Breadcrumb
        items={[
          {
            label: __("Statements of Applicability"),
            to: listUrl,
          },
          {
            label:
                            statementOfApplicability.name
                            || __("Statement of Applicability detail"),
          },
        ]}
      />

      <PageHeader
        title={
          isEditingName && canUpdate
            ? (
                <div className="flex items-center gap-2">
                  <Input
                    {...registerName("name")}
                    variant="title"
                    className="flex-1"
                    autoFocus
                    onKeyDown={(e) => {
                      if (e.key === "Escape") {
                        handleCancelNameEdit();
                      }
                      if (e.key === "Enter" && e.ctrlKey) {
                        void handleUpdateName();
                      }
                    }}
                  />
                  <Button
                    variant="quaternary"
                    icon={IconCheckmark1}
                    onClick={() => void handleUpdateName()}
                    disabled={isUpdating}
                  />
                  <Button
                    variant="quaternary"
                    icon={IconCrossLargeX}
                    onClick={handleCancelNameEdit}
                  />
                </div>
              )
            : (
                <div className="flex items-center gap-2">
                  <span>{statementOfApplicability.name || ""}</span>
                  {canUpdate && (
                    <Button
                      variant="quaternary"
                      icon={IconPencil}
                      onClick={() => setIsEditingName(true)}
                    />
                  )}
                </div>
              )
        }
      >
        {statementOfApplicability.canExport && (
          <Button
            variant="secondary"
            icon={IconArrowDown}
            onClick={() => void handleExport()}
            disabled={isExporting}
          >
            {__("Export")}
          </Button>
        )}
        {canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>

      <div className="space-y-6">
        <div className="space-y-4">
          <h2 className="text-base font-medium">{__("Details")}</h2>
          <Card className="space-y-4" padded>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Owner")}
                </div>
                {isEditingOwner && canUpdate
                  ? (
                      <div className="flex items-center gap-2">
                        <Suspense
                          fallback={
                            <div>{__("Loading...")}</div>
                          }
                        >
                          <PeopleSelectField
                            organizationId={organizationId}
                            control={controlOwner}
                            name="ownerId"
                          />
                        </Suspense>
                        <Button
                          variant="quaternary"
                          icon={IconCheckmark1}
                          onClick={() => void handleUpdateOwner()}
                          disabled={isUpdating}
                        />
                        <Button
                          variant="quaternary"
                          icon={IconCrossLargeX}
                          onClick={handleCancelOwnerEdit}
                        />
                      </div>
                    )
                  : (
                      <div className="flex items-center gap-2">
                        <div className="text-sm text-txt-primary">
                          {statementOfApplicability.owner
                            ?.fullName || "-"}
                        </div>
                        {canUpdate && (
                          <Button
                            variant="quaternary"
                            icon={IconPencil}
                            onClick={() =>
                              setIsEditingOwner(true)}
                          />
                        )}
                      </div>
                    )}
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Created at")}
                </div>
                <div className="text-sm text-txt-primary">
                  {formatDate(statementOfApplicability.createdAt)}
                </div>
              </div>
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Updated at")}
                </div>
                <div className="text-sm text-txt-primary">
                  {formatDate(statementOfApplicability.updatedAt)}
                </div>
              </div>
            </div>
          </Card>
        </div>

        {statementOfApplicability.id && (
          <div className="space-y-4">
            <h2 className="text-base font-medium">
              {__("Statements")}
            </h2>
            <StatementOfApplicabilityControlsTab
              statementOfApplicability={
                statementOfApplicability as typeof statementOfApplicability & {
                  id: string;
                }
              }
              isSnapshotMode={isSnapshotMode}
            />
          </div>
        )}
      </div>
    </div>
  );
}
