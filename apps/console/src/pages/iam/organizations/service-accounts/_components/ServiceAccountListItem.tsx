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

import { dateFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ServiceAccountListItem_serviceAccount$key } from "#/__generated__/iam/ServiceAccountListItem_serviceAccount.graphql";
import type { ServiceAccountListItemDeleteMutation } from "#/__generated__/iam/ServiceAccountListItemDeleteMutation.graphql";
import type { ServiceAccountListItemDisableMutation } from "#/__generated__/iam/ServiceAccountListItemDisableMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { formatAPIScopeLabel } from "#/pages/iam/oauthTokens/_components/scopeLabels";

import { ConfirmServiceAccountActionDialog } from "./ConfirmServiceAccountActionDialog";
import { ServiceAccountCredentialsDialog } from "./ServiceAccountCredentialsDialog";
import { UpdateServiceAccountDialog } from "./UpdateServiceAccountDialog";

const serviceAccountListItemFragment = graphql`
  fragment ServiceAccountListItem_serviceAccount on ServiceAccount {
    id
    name
    scopes
    disabledAt
    createdAt
    canUpdate: permission(action: "iam:service-account:update")
    canDisable: permission(action: "iam:service-account:disable")
    canDelete: permission(action: "iam:service-account:delete")
    canListCredentials: permission(
      action: "iam:service-account-credential:list"
    )
    ...ServiceAccountCredentialsDialog_serviceAccount
    ...UpdateServiceAccountDialog_serviceAccount
  }
`;

const disableServiceAccountMutation = graphql`
  mutation ServiceAccountListItemDisableMutation(
    $input: DisableServiceAccountInput!
  ) {
    disableServiceAccount(input: $input) {
      serviceAccount {
        ...ServiceAccountListItem_serviceAccount
      }
    }
  }
`;

const deleteServiceAccountMutation = graphql`
  mutation ServiceAccountListItemDeleteMutation(
    $input: DeleteServiceAccountInput!
    $connections: [ID!]!
  ) {
    deleteServiceAccount(input: $input) {
      deletedServiceAccountId @deleteEdge(connections: $connections)
    }
  }
`;

interface ServiceAccountListItemProps {
  connectionId: string;
  serviceAccountKey: ServiceAccountListItem_serviceAccount$key;
}

export function ServiceAccountListItem(
  { connectionId, serviceAccountKey }: ServiceAccountListItemProps,
) {
  const { t, i18n } = useTranslation("iam/organizations/service-accounts");
  const serviceAccount = useFragment(
    serviceAccountListItemFragment,
    serviceAccountKey,
  );
  const [updateOpen, setUpdateOpen] = useState(false);
  const [credentialsOpen, setCredentialsOpen] = useState(false);
  const [disableOpen, setDisableOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [disableServiceAccount, isDisabling]
    = useMutation<ServiceAccountListItemDisableMutation>(
      disableServiceAccountMutation,
      {
        successMessage: t("messages.disabled"),
        errorToast: t("errors.disable"),
      },
    );
  const [deleteServiceAccount, isDeleting]
    = useMutation<ServiceAccountListItemDeleteMutation>(
      deleteServiceAccountMutation,
      {
        successMessage: t("messages.deleted"),
        errorToast: t("errors.delete"),
      },
    );

  async function disable() {
    try {
      await disableServiceAccount({
        variables: {
          input: { serviceAccountId: serviceAccount.id },
        },
      });
      setDisableOpen(false);
    } catch {
      return;
    }
  }

  async function remove() {
    try {
      await deleteServiceAccount({
        variables: {
          input: { serviceAccountId: serviceAccount.id },
          connections: [connectionId],
        },
      });
      setDeleteOpen(false);
    } catch {
      return;
    }
  }

  const isDisabled = serviceAccount.disabledAt != null;
  const visibleScopes = serviceAccount.scopes.slice(0, 3);
  const extraScopeCount = serviceAccount.scopes.length - visibleScopes.length;

  return (
    <>
      <TableRow>
        <TableCell>
          <div className="flex flex-col gap-1">
            <span className="font-medium text-sand-12">
              {serviceAccount.name}
            </span>
            <span className="text-1 text-sand-10">
              {serviceAccount.id}
            </span>
          </div>
        </TableCell>
        <TableCell>
          <Badge
            color={isDisabled ? "red" : "green"}
            variant="soft"
          >
            {isDisabled ? t("status.disabled") : t("status.active")}
          </Badge>
        </TableCell>
        <TableCell>
          <div className="flex flex-wrap gap-1">
            {visibleScopes.map(scope => (
              <Badge key={scope} size={1} variant="soft">
                {formatAPIScopeLabel(scope, t)}
              </Badge>
            ))}
            {extraScopeCount > 0 && (
              <Badge size={1} variant="outline">
                {t("values.moreScopes", { count: extraScopeCount })}
              </Badge>
            )}
          </div>
        </TableCell>
        <TableCell>
          {dateFormat(i18n.language, serviceAccount.createdAt)}
        </TableCell>
        <TableCell interactive>
          <div className="flex flex-wrap justify-end gap-2">
            {!isDisabled && serviceAccount.canListCredentials && (
              <Button
                size={1}
                variant="soft"
                onClick={() => setCredentialsOpen(true)}
              >
                {t("actions.credentials")}
              </Button>
            )}
            {!isDisabled && serviceAccount.canUpdate && (
              <Button
                size={1}
                variant="soft"
                onClick={() => setUpdateOpen(true)}
              >
                {t("actions.edit")}
              </Button>
            )}
            {!isDisabled && serviceAccount.canDisable && (
              <Button
                size={1}
                variant="soft"
                color="amber"
                onClick={() => setDisableOpen(true)}
              >
                {t("actions.disable")}
              </Button>
            )}
            {serviceAccount.canDelete && (
              <Button
                size={1}
                variant="soft"
                color="red"
                onClick={() => setDeleteOpen(true)}
              >
                {t("actions.delete")}
              </Button>
            )}
          </div>
        </TableCell>
      </TableRow>

      <UpdateServiceAccountDialog
        open={updateOpen}
        serviceAccountKey={serviceAccount}
        onOpenChange={setUpdateOpen}
      />
      <ServiceAccountCredentialsDialog
        open={credentialsOpen}
        serviceAccountKey={serviceAccount}
        onOpenChange={setCredentialsOpen}
      />
      <ConfirmServiceAccountActionDialog
        cancelLabel={t("actions.cancel")}
        confirmLabel={t("actions.disable")}
        description={t("disable.description", { name: serviceAccount.name })}
        isSubmitting={isDisabling}
        open={disableOpen}
        title={t("disable.title")}
        onConfirm={() => void disable()}
        onOpenChange={setDisableOpen}
      />
      <ConfirmServiceAccountActionDialog
        cancelLabel={t("actions.cancel")}
        confirmLabel={t("actions.delete")}
        description={t("delete.description", { name: serviceAccount.name })}
        isSubmitting={isDeleting}
        open={deleteOpen}
        title={t("delete.title")}
        onConfirm={() => void remove()}
        onOpenChange={setDeleteOpen}
      />
    </>
  );
}
