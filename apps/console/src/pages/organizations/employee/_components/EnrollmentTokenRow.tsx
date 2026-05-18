// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { formatDate, formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Td, Tr, useConfirm, useToast } from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { EnrollmentTokenRowFragment$key } from "#/__generated__/core/EnrollmentTokenRowFragment.graphql";
import type { EnrollmentTokenRowRevokeMutation } from "#/__generated__/core/EnrollmentTokenRowRevokeMutation.graphql";

const enrollmentTokenRowFragment = graphql`
  fragment EnrollmentTokenRowFragment on DeviceEnrollmentToken {
    id
    name
    createdAt
    expiresAt
    revokedAt
    maxUses
    usedCount
  }
`;

const revokeEnrollmentTokenMutation = graphql`
  mutation EnrollmentTokenRowRevokeMutation(
    $input: RevokeDeviceEnrollmentTokenInput!
  ) {
    revokeDeviceEnrollmentToken(input: $input) {
      enrollmentToken {
        id
        revokedAt
      }
    }
  }
`;

interface EnrollmentTokenRowProps {
  fKey: EnrollmentTokenRowFragment$key;
}

export function EnrollmentTokenRow({ fKey }: EnrollmentTokenRowProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();

  const token = useFragment(enrollmentTokenRowFragment, fKey);

  const [revokeToken] = useMutation<EnrollmentTokenRowRevokeMutation>(
    revokeEnrollmentTokenMutation,
  );

  const isRevoked = Boolean(token.revokedAt);

  const handleRevoke = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          revokeToken({
            variables: { input: { enrollmentTokenId: token.id } },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: errors[0].message,
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Enrollment token revoked"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to revoke enrollment token"),
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
          "Revoke this enrollment token? Devices using it can finish their current enrollment but no new devices will be able to use it.",
        ),
        variant: "danger",
        label: __("Revoke"),
      },
    );
  };

  return (
    <Tr>
      <Td>{token.name}</Td>
      <Td>{formatDate(token.createdAt)}</Td>
      <Td>{formatDate(token.expiresAt)}</Td>
      <Td>
        {token.usedCount}
        {" / "}
        {token.maxUses ?? "∞"}
      </Td>
      <Td className="text-end">
        {isRevoked
          ? <span className="text-tertiary text-xs">{__("Revoked")}</span>
          : (
              <Button variant="secondary" onClick={handleRevoke}>
                {__("Revoke")}
              </Button>
            )}
      </Td>
    </Tr>
  );
}
