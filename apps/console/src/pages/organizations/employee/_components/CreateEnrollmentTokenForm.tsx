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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Input, useToast } from "@probo/ui";
import { useState } from "react";
import { useMutation } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { CreateEnrollmentTokenFormMutation } from "#/__generated__/core/CreateEnrollmentTokenFormMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { EnrollmentInstructions } from "./EnrollmentInstructions";

const TOKEN_VALIDITY_SECONDS = 60 * 60 * 24 * 7;
const TOKEN_DEFAULT_MAX_USES = 25;

const createEnrollmentTokenMutation = graphql`
  mutation CreateEnrollmentTokenFormMutation(
    $input: CreateDeviceEnrollmentTokenInput!
    $connections: [ID!]!
  ) {
    createDeviceEnrollmentToken(input: $input) {
      secret
      enrollmentToken
        @prependNode(
          connections: $connections
          edgeTypeName: "DeviceEnrollmentTokenEdge"
        ) {
        id
        name
        createdAt
        expiresAt
        revokedAt
        maxUses
        usedCount
      }
    }
  }
`;

interface CreateEnrollmentTokenFormProps {
  connectionId: DataID;
}

export function CreateEnrollmentTokenForm(
  { connectionId }: CreateEnrollmentTokenFormProps,
) {
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [tokenName, setTokenName] = useState("");
  const [secret, setSecret] = useState<string | null>(null);

  const organizationId = useOrganizationId();
  const [createEnrollmentToken, isCreating]
    = useMutation<CreateEnrollmentTokenFormMutation>(
      createEnrollmentTokenMutation,
    );

  const handleCreate = () => {
    const name = tokenName.trim();
    if (!name) return;

    createEnrollmentToken({
      variables: {
        input: {
          organizationId,
          name,
          validitySeconds: TOKEN_VALIDITY_SECONDS,
          maxUses: TOKEN_DEFAULT_MAX_USES,
        },
        connections: [connectionId],
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }
        setSecret(response.createDeviceEnrollmentToken.secret);
        setTokenName("");
        toast({
          title: __("Success"),
          description: __(
            "Enrollment token created. Copy it now — it will not be shown again.",
          ),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create enrollment token"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <Input
          placeholder={__("Token name (e.g. \"My laptop\")")}
          value={tokenName}
          onChange={e => setTokenName(e.target.value)}
          className="w-72"
        />
        <Button
          onClick={handleCreate}
          disabled={isCreating || !tokenName.trim()}
        >
          {__("Generate enrollment token")}
        </Button>
      </div>
      {secret && <EnrollmentInstructions secret={secret} />}
    </div>
  );
}
