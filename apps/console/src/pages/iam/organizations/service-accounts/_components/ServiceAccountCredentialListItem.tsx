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

import type { ServiceAccountCredentialListItem_credential$key } from "#/__generated__/iam/ServiceAccountCredentialListItem_credential.graphql";
import type { ServiceAccountCredentialListItem_serviceAccount$key } from "#/__generated__/iam/ServiceAccountCredentialListItem_serviceAccount.graphql";
import type { ServiceAccountCredentialListItemRevokeMutation } from "#/__generated__/iam/ServiceAccountCredentialListItemRevokeMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { ConfirmServiceAccountActionDialog } from "./ConfirmServiceAccountActionDialog";

const credentialFragment = graphql`
  fragment ServiceAccountCredentialListItem_credential
  on ServiceAccountCredential {
    id
    name
    expiresAt
    lastUsedAt
    revokedAt
  }
`;

const serviceAccountFragment = graphql`
  fragment ServiceAccountCredentialListItem_serviceAccount
  on ServiceAccount {
    id
    canRevokeCredential: permission(
      action: "iam:service-account-credential:revoke"
    )
  }
`;

const revokeCredentialMutation = graphql`
  mutation ServiceAccountCredentialListItemRevokeMutation(
    $input: RevokeServiceAccountCredentialInput!
  ) {
    revokeServiceAccountCredential(input: $input) {
      serviceAccountCredential {
        ...ServiceAccountCredentialListItem_credential
      }
    }
  }
`;

interface ServiceAccountCredentialListItemProps {
  credentialKey: ServiceAccountCredentialListItem_credential$key;
  serviceAccountKey: ServiceAccountCredentialListItem_serviceAccount$key;
}

export function ServiceAccountCredentialListItem(
  props: ServiceAccountCredentialListItemProps,
) {
  const { credentialKey, serviceAccountKey } = props;
  const { t, i18n } = useTranslation("iam/organizations/service-accounts");
  const credential = useFragment(credentialFragment, credentialKey);
  const serviceAccount = useFragment(
    serviceAccountFragment,
    serviceAccountKey,
  );
  const [revokeOpen, setRevokeOpen] = useState(false);
  const [revokeCredential, isRevoking]
    = useMutation<ServiceAccountCredentialListItemRevokeMutation>(
      revokeCredentialMutation,
      {
        successMessage: t("messages.credentialRevoked"),
        errorToast: t("errors.revokeCredential"),
      },
    );

  async function revoke() {
    try {
      await revokeCredential({
        variables: {
          input: {
            serviceAccountId: serviceAccount.id,
            serviceAccountCredentialId: credential.id,
          },
        },
      });
      setRevokeOpen(false);
    } catch {
      return;
    }
  }

  const isRevoked = credential.revokedAt != null;

  return (
    <>
      <TableRow>
        <TableCell>{credential.name}</TableCell>
        <TableCell>
          <Badge
            color={isRevoked ? "red" : "green"}
            variant="soft"
          >
            {isRevoked ? t("status.revoked") : t("status.active")}
          </Badge>
        </TableCell>
        <TableCell>{dateFormat(i18n.language, credential.expiresAt)}</TableCell>
        <TableCell>
          {credential.lastUsedAt
            ? dateFormat(i18n.language, credential.lastUsedAt)
            : t("values.never")}
        </TableCell>
        <TableCell interactive>
          {!isRevoked && serviceAccount.canRevokeCredential && (
            <Button
              size={1}
              variant="soft"
              color="red"
              loading={isRevoking}
              onClick={() => setRevokeOpen(true)}
            >
              {t("actions.revoke")}
            </Button>
          )}
        </TableCell>
      </TableRow>
      <ConfirmServiceAccountActionDialog
        cancelLabel={t("actions.cancel")}
        confirmLabel={t("actions.revoke")}
        description={t("credentials.revokeDescription", {
          name: credential.name,
        })}
        isSubmitting={isRevoking}
        open={revokeOpen}
        title={t("credentials.revokeTitle")}
        onConfirm={() => void revoke()}
        onOpenChange={setRevokeOpen}
      />
    </>
  );
}
