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

import {
  formatDate,
  formatError,
  getCloudAccountProviderLabel,
  type GraphQLError,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  DropdownItem,
  IconChevronDown,
  IconPlusLarge,
  IconRotateCw,
  IconTrashCan,
  Spinner,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { graphql, useMutation, usePaginationFragment } from "react-relay";

import type { CloudAccountListDeleteMutation } from "#/__generated__/core/CloudAccountListDeleteMutation.graphql";
import type {
  CloudAccountListFragment$data,
  CloudAccountListFragment$key,
} from "#/__generated__/core/CloudAccountListFragment.graphql";
import type { CloudAccountListPaginationQuery } from "#/__generated__/core/CloudAccountListPaginationQuery.graphql";
import type { CloudAccountListVerifyMutation } from "#/__generated__/core/CloudAccountListVerifyMutation.graphql";

import { CloudAccountConnectDialog } from "./CloudAccountConnectDialog";
import { CloudAccountReconnectDialog } from "./CloudAccountReconnectDialog";
import { CloudAccountStatusBadge } from "./CloudAccountStatusBadge";

const fragment = graphql`
  fragment CloudAccountListFragment on Organization
  @refetchable(queryName: "CloudAccountListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 25 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    canCreate: permission(action: "core:cloud-account:create")
    cloudAccounts(first: $first, after: $after)
      @connection(key: "CloudAccountList_cloudAccounts") {
      __id
      edges {
        node {
          id
          provider
          label
          status
          scope {
            kind
            identifier
          }
          lastVerifiedAt
          canRotate: permission(action: "core:cloud-account:rotate-credentials")
          canVerify: permission(action: "core:cloud-account:verify")
          canDelete: permission(action: "core:cloud-account:delete")
        }
      }
    }
  }
`;

const verifyMutation = graphql`
  mutation CloudAccountListVerifyMutation($input: VerifyCloudAccountInput!) {
    verifyCloudAccount(input: $input) {
      cloudAccount {
        id
        status
        lastVerifiedAt
      }
      status
      lastProbeError
    }
  }
`;

const deleteMutation = graphql`
  mutation CloudAccountListDeleteMutation(
    $input: DeleteCloudAccountInput!
    $connections: [ID!]!
  ) {
    deleteCloudAccount(input: $input) {
      deletedCloudAccountId @deleteEdge(connections: $connections)
    }
  }
`;

type Props = {
  fKey: CloudAccountListFragment$key;
};

type CloudAccountNode =
  CloudAccountListFragment$data["cloudAccounts"]["edges"][number]["node"];

const SCOPE_IDENTIFIER_MAX = 24;

function truncateIdentifier(identifier: string | null | undefined): string {
  if (!identifier) return "—";
  if (identifier.length <= SCOPE_IDENTIFIER_MAX) return identifier;
  return `${identifier.slice(0, SCOPE_IDENTIFIER_MAX)}…`;
}

export function CloudAccountList(props: Props) {
  const { fKey } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const connectDialogRef = useDialogRef();
  const reconnectDialogRef = useDialogRef();
  const [reconnectTarget, setReconnectTarget] =
    useState<CloudAccountNode | null>(null);

  const { data, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    CloudAccountListPaginationQuery,
    CloudAccountListFragment$key
  >(fragment, fKey);

  const cloudAccounts = data.cloudAccounts.edges.map((edge) => edge.node);
  const canCreate = data.canCreate;
  const connectionId = data.cloudAccounts.__id;
  const hasAnyAction = cloudAccounts.some(
    (account) => account.canRotate || account.canVerify || account.canDelete,
  );

  const [verifyCloudAccount, isVerifying] =
    useMutation<CloudAccountListVerifyMutation>(verifyMutation);
  const [deleteCloudAccount] =
    useMutation<CloudAccountListDeleteMutation>(deleteMutation);

  const handleConnectCloud = () => {
    connectDialogRef.current?.open();
  };

  const handleVerify = (account: CloudAccountNode) => {
    verifyCloudAccount({
      variables: { input: { cloudAccountId: account.id } },
      onCompleted(response, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors.map((e) => e.message).join(", "),
            variant: "error",
          });
          return;
        }
        const status = response.verifyCloudAccount.status;
        if (status === "VERIFIED") {
          toast({
            title: __("Verified"),
            description: __("Cloud account verified successfully."),
            variant: "success",
          });
        } else {
          toast({
            title: __("Verification failed"),
            description:
              response.verifyCloudAccount.lastProbeError ??
              __("The probe did not succeed."),
            variant: "error",
          });
        }
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Could not verify cloud account"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleDelete = (account: CloudAccountNode) => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteCloudAccount({
            variables: {
              input: { cloudAccountId: account.id },
              connections: [connectionId],
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: errors.map((e) => e.message).join(", "),
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Deleted"),
                  description: __("Cloud account deleted."),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Could not delete cloud account"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: __(
          "Are you sure you want to delete this cloud account? Audits relying on it will stop working.",
        ),
        label: __("Delete"),
      },
    );
  };

  const handleRotate = (account: CloudAccountNode) => {
    setReconnectTarget(account);
    reconnectDialogRef.current?.open();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        {canCreate && (
          <Button icon={IconPlusLarge} onClick={handleConnectCloud}>
            {__("Connect a cloud")}
          </Button>
        )}
      </div>
      {cloudAccounts.length === 0 ? (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No cloud accounts yet")}
            </h3>
            {canCreate && (
              <p className="text-txt-tertiary">
                {__(
                  "Connect AWS, GCP, or Azure to begin running cloud-aware audits.",
                )}
              </p>
            )}
          </div>
        </Card>
      ) : (
        <Table>
          <Thead>
            <Tr>
              <Th>{__("Provider")}</Th>
              <Th>{__("Label")}</Th>
              <Th>{__("Scope")}</Th>
              <Th>{__("Status")}</Th>
              <Th>{__("Last verified")}</Th>
              {hasAnyAction && <Th className="w-18" />}
            </Tr>
          </Thead>
          <Tbody>
            {cloudAccounts.map((account) => {
              const rowHasAction =
                account.canRotate || account.canVerify || account.canDelete;
              return (
                <Tr key={account.id}>
                  <Td>
                    <Badge variant="neutral">
                      {getCloudAccountProviderLabel(__, account.provider)}
                    </Badge>
                  </Td>
                  <Td>{account.label}</Td>
                  <Td>
                    <div className="flex flex-col gap-0.5">
                      <span className="text-xs text-txt-tertiary">
                        {account.scope.kind}
                      </span>
                      <span
                        className="text-sm font-mono"
                        title={account.scope.identifier ?? undefined}
                      >
                        {truncateIdentifier(account.scope.identifier)}
                      </span>
                    </div>
                  </Td>
                  <Td>
                    <CloudAccountStatusBadge status={account.status} />
                  </Td>
                  <Td>
                    {account.lastVerifiedAt
                      ? formatDate(account.lastVerifiedAt)
                      : "—"}
                  </Td>
                  {hasAnyAction && (
                    <Td noLink width={50} className="text-end w-18">
                      {rowHasAction && (
                        <ActionDropdown>
                          {account.canVerify && (
                            <DropdownItem
                              icon={IconRotateCw}
                              disabled={isVerifying}
                              onClick={() => handleVerify(account)}
                            >
                              {__("Verify now")}
                            </DropdownItem>
                          )}
                          {account.canRotate && (
                            <DropdownItem
                              icon={IconRotateCw}
                              onClick={() => handleRotate(account)}
                            >
                              {__("Reconnect")}
                            </DropdownItem>
                          )}
                          {account.canDelete && (
                            <DropdownItem
                              variant="danger"
                              icon={IconTrashCan}
                              onClick={() => handleDelete(account)}
                            >
                              {__("Delete")}
                            </DropdownItem>
                          )}
                        </ActionDropdown>
                      )}
                    </Td>
                  )}
                </Tr>
              );
            })}
          </Tbody>
        </Table>
      )}
      {hasNext && (
        <Button
          variant="tertiary"
          onClick={() => loadNext(25)}
          disabled={isLoadingNext}
          className="mt-3 mx-auto"
          icon={IconChevronDown}
        >
          {isLoadingNext && <Spinner />}
          {__("Show more")}
        </Button>
      )}
      <CloudAccountConnectDialog
        ref={connectDialogRef}
        connectionId={connectionId}
      />
      {reconnectTarget && (
        <CloudAccountReconnectDialog
          ref={reconnectDialogRef}
          cloudAccountId={reconnectTarget.id}
          provider={reconnectTarget.provider}
        />
      )}
    </div>
  );
}
