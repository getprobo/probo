import { formatDate, formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Spinner, Td, Tr, useConfirm, useToast } from "@probo/ui";
import { clsx } from "clsx";
import { Suspense } from "react";
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
    token @include(if: $includeToken)
  }
`;

export function PersonalAPIKeyRow(props: {
  fKey: PersonalAPIKeyRowFragment$key;
  connectionId: string;
}) {
  const { fKey, connectionId } = props;
  const { __ } = useTranslate();
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
                  title: __("Error"),
                  description: formatError(
                    __("Failed to revoke API key."),
                    errors as GraphQLError[],
                  ),
                  variant: "error",
                });
                reject(new Error(errors[0]?.message ?? __("Failed to revoke API key.")));
                return;
              }
              toast({
                title: __("Success"),
                description: __("API key revoked successfully."),
                variant: "success",
              });
              resolve();
            },
            onError: (error) => {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to revoke API key."),
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
        title: __("Revoke API Key"),
        message: __(
          `Are you sure you want to revoke the API key "${key.name}"? This action cannot be undone.`,
        ),
        label: __("Revoke"),
        variant: "danger",
      },
    );
  };

  return (
    <Tr className={clsx(isRevoking && "opacity-60 pointer-events-none")}>
      <Td>
        <div className="font-medium text-txt-primary">{key.name}</div>
        <div className="text-xs text-txt-tertiary">
          {expired ? __("Expired") : __("Active")}
        </div>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {key.lastUsedAt ? formatDate(key.lastUsedAt) : __("Never")}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {formatDate(key.createdAt)}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {formatDate(key.expiresAt)}
        </span>
      </Td>
      <Td width={140} className="text-end">
        <div className="flex gap-2 justify-end">
          <Suspense fallback={<Spinner />}>
            <PersonalAPIKeyTokenAction fKey={fKey} disabled={isRevoking} />
          </Suspense>
          <Button variant="danger" onClick={handleRevoke} disabled={isRevoking}>
            {__("Revoke")}
          </Button>
        </div>
      </Td>
    </Tr>
  );
}
