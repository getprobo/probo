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

import { formatDate, formatError, type GraphQLError, promisifyMutation, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  IconCheckmark1,
  IconCrossLargeX,
  IconPencil,
  IconTrashCan,
  IconUpload,
  Input,
  PageHeader,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useState } from "react";
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
import type { StatementOfApplicabilityDetailPageQuery } from "#/__generated__/core/StatementOfApplicabilityDetailPageQuery.graphql";
import type { StatementOfApplicabilityDetailPageUpdateMutation } from "#/__generated__/core/StatementOfApplicabilityDetailPageUpdateMutation.graphql";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { PublishStatementOfApplicabilityDialog } from "./dialogs/PublishStatementOfApplicabilityDialog";
import StatementOfApplicabilityControlsTab from "./tabs/StatementOfApplicabilityControlsTab";

export const statementOfApplicabilityDetailPageQuery = graphql`
    query StatementOfApplicabilityDetailPageQuery($statementOfApplicabilityId: ID!) {
        node(id: $statementOfApplicabilityId) {
            ... on StatementOfApplicability {
                id
                name
                createdAt
                updatedAt
                defaultApprovers {
                    id
                    fullName
                }
                canUpdate: permission(action: "core:statement-of-applicability:update")
                canDelete: permission(action: "core:statement-of-applicability:delete")
                canPublish: permission(action: "core:statement-of-applicability:publish")
                ...StatementOfApplicabilityControlsTabFragment
            }
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
                createdAt
                updatedAt
                defaultApprovers {
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
  const { statementOfApplicabilityId } = useParams<{
    statementOfApplicabilityId: string;
  }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(statementOfApplicabilityDetailPageQuery, props.queryRef);
  const statementOfApplicability = data.node;
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const { toast } = useToast();

  if (!statementOfApplicabilityId || !statementOfApplicability) {
    throw new Error(
      "Cannot load statement of applicability detail page without statementOfApplicabilityId parameter",
    );
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    StatementOfApplicabilityConnectionKey,
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
  const [updateStatementOfApplicability, isUpdating]
    = useMutation<StatementOfApplicabilityDetailPageUpdateMutation>(updateMutation);

  const canUpdate = statementOfApplicability.canUpdate;
  const canDelete = statementOfApplicability.canDelete;

  const nameSchema = z.object({
    name: z.string().min(1, __("Name is required")),
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

  const handleUpdateName = handleSubmitName((data) => {
    if (!statementOfApplicability.id) return;

    updateStatementOfApplicability({
      variables: {
        input: {
          id: statementOfApplicability.id,
          name: data.name,
        },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Statement of Applicability updated successfully."),
          variant: "success",
        });
        setIsEditingName(false);
        resetName({ name: data.name });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to update Statement of Applicability"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  });

  const handleCancelNameEdit = () => {
    setIsEditingName(false);
    resetName({
      name: statementOfApplicability.name || "",
    });
  };

  const [isEditingApprovers, setIsEditingApprovers] = useState(false);
  const defaultApproverIds = (statementOfApplicability.defaultApprovers ?? []).map(a => a.id);

  const approversSchema = z.object({
    defaultApproverIds: z.array(z.string()),
  });

  const {
    control: approversControl,
    handleSubmit: handleSubmitApprovers,
    reset: resetApprovers,
  } = useFormWithSchema(approversSchema, {
    defaultValues: {
      defaultApproverIds,
    },
  });

  const handleUpdateApprovers = handleSubmitApprovers((data) => {
    if (!statementOfApplicability.id) return;

    updateStatementOfApplicability({
      variables: {
        input: {
          id: statementOfApplicability.id,
          defaultApproverIds: data.defaultApproverIds,
        },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Default approvers updated successfully."),
          variant: "success",
        });
        setIsEditingApprovers(false);
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to update default approvers"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  });

  const handleCancelApproversEdit = () => {
    setIsEditingApprovers(false);
    resetApprovers({ defaultApproverIds });
  };

  const listUrl = `/organizations/${organizationId}/statements-of-applicability`;

  return (
    <div className="space-y-6">
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
        {statementOfApplicability.canPublish && statementOfApplicability.id && (
          <PublishStatementOfApplicabilityDialog
            statementOfApplicabilityId={statementOfApplicability.id}
            defaultApproverIds={defaultApproverIds}
          >
            <Button
              variant="secondary"
              icon={IconUpload}
            >
              {__("Publish")}
            </Button>
          </PublishStatementOfApplicabilityDialog>
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
            <div>
              <div className="flex items-center justify-between mb-1">
                <div className="text-xs text-txt-tertiary font-semibold">
                  {__("Default Approvers")}
                </div>
                {canUpdate && !isEditingApprovers && (
                  <Button
                    variant="quaternary"
                    icon={IconPencil}
                    onClick={() => setIsEditingApprovers(true)}
                  />
                )}
              </div>
              {isEditingApprovers && canUpdate
                ? (
                    <div className="space-y-2">
                      <PeopleMultiSelectField
                        name="defaultApproverIds"
                        control={approversControl}
                        organizationId={organizationId}
                        placeholder={__("Add approvers...")}
                      />
                      <div className="flex gap-2">
                        <Button
                          variant="quaternary"
                          icon={IconCheckmark1}
                          onClick={() => void handleUpdateApprovers()}
                          disabled={isUpdating}
                        />
                        <Button
                          variant="quaternary"
                          icon={IconCrossLargeX}
                          onClick={handleCancelApproversEdit}
                        />
                      </div>
                    </div>
                  )
                : (
                    <div className="text-sm text-txt-primary">
                      {statementOfApplicability.defaultApprovers
                        && statementOfApplicability.defaultApprovers.length > 0
                        ? statementOfApplicability.defaultApprovers
                            .map(a => a.fullName)
                            .join(", ")
                        : __("None")}
                    </div>
                  )}
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
            />
          </div>
        )}
      </div>
    </div>
  );
}
