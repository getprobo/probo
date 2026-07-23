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

import { formatError } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import { Button, Spinner, Td, Tr, useConfirm, useToast } from "@probo/ui";
import { clsx } from "clsx";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { PersonalAPIKeyRow_revokeMutation } from "#/__generated__/iam/PersonalAPIKeyRow_revokeMutation.graphql";
import type { PersonalAPIKeyRowFragment$key } from "#/__generated__/iam/PersonalAPIKeyRowFragment.graphql";

import { PersonalAPIKeyTokenAction } from "./PersonalAPIKeyTokenAction";

const revokeMutation = graphql`
  mutation PersonalAPIKeyRow_revokeMutation(
    $input: RevokePersonalAPIKeyInput!
    $connections: [ID!]!
  ) {
    revokePersonalAPIKey(input: $input) {
      personalAPIKeyId @deleteEdge(connections: $connections)
    }
  }
`;

export const personalAPIKeyRowFragment = graphql`
  fragment PersonalAPIKeyRowFragment on PersonalAPIKey
  @refetchable(queryName: "PersonalAPIKeyRowRefetchQuery")
  @argumentDefinitions(includeToken: { type: "Boolean", defaultValue: false }) {
    id
    name
    createdAt
    expiresAt
    lastUsedAt
    # eslint-disable-next-line relay/unused-fields
    token @include(if: $includeToken)
  }
`;

export function PersonalAPIKeyRow(props: {
  fKey: PersonalAPIKeyRowFragment$key;
  connectionId: string;
}) {
  const { fKey, connectionId } = props;
  const { t, i18n } = useTranslation();
  const confirm = useConfirm();
  const { toast } = useToast();
  const now = new Date();

  const key = useFragment(personalAPIKeyRowFragment, fKey);
  const expired = new Date(key.expiresAt) < now;

  const [revokeCommit, isRevoking]
    = useMutation<PersonalAPIKeyRow_revokeMutation>(revokeMutation);

  const handleRevoke = () => {
    confirm(
      async () => {
        await new Promise<void>((resolve, reject) => {
          revokeCommit({
            variables: {
              input: { personalAPIKeyId: key.id },
              connections: [connectionId],
            },
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                toast({
                  title: t("common.error"),
                  description: formatError(
                    t("personalApiKeyRow.errors.revoke"),
                    errors,
                  ),
                  variant: "error",
                });
                reject(new Error(errors[0]?.message ?? t("personalApiKeyRow.errors.revoke")));
                return;
              }
              toast({
                title: t("common.success"),
                description: t("personalApiKeyRow.messages.revoked"),
                variant: "success",
              });
              resolve();
            },
            onError: (error) => {
              toast({
                title: t("common.error"),
                description: formatError(
                  t("personalApiKeyRow.errors.revoke"),
                  error,
                ),
                variant: "error",
              });
              reject(error);
            },
          });
        });
      },
      {
        title: t("personalApiKeyRow.revoke.title"),
        message: t("personalApiKeyRow.revoke.confirmation", { name: key.name }),
        label: t("personalApiKeyRow.actions.revoke"),
        variant: "danger",
      },
    );
  };

  return (
    <Tr className={clsx(isRevoking && "opacity-60 pointer-events-none")}>
      <Td>
        <div className="font-medium text-txt-primary">{key.name}</div>
        <div className="text-xs text-txt-tertiary">
          {expired ? t("personalApiKeyRow.status.expired") : t("personalApiKeyRow.status.active")}
        </div>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {key.lastUsedAt ? dateFormat(i18n.language, key.lastUsedAt) : t("personalApiKeyRow.never")}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {dateFormat(i18n.language, key.createdAt)}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {dateFormat(i18n.language, key.expiresAt)}
        </span>
      </Td>
      <Td width={140} className="text-end">
        <div className="flex gap-2 justify-end">
          <Suspense fallback={<Spinner />}>
            <PersonalAPIKeyTokenAction fKey={fKey} disabled={isRevoking} />
          </Suspense>
          <Button variant="danger" onClick={handleRevoke} disabled={isRevoking}>
            {t("personalApiKeyRow.actions.revoke")}
          </Button>
        </div>
      </Td>
    </Tr>
  );
}
