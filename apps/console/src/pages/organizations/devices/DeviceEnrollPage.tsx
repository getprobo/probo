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

import { formatDate, promisifyMutation } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Input,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { DeviceGraphEnrollmentTokensQuery } from "#/__generated__/core/DeviceGraphEnrollmentTokensQuery.graphql";
import {
  deviceEnrollmentTokensQuery,
  useCreateDeviceEnrollmentTokenMutation,
  useRevokeDeviceEnrollmentToken,
} from "#/hooks/graph/DeviceGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  queryRef: PreloadedQuery<DeviceGraphEnrollmentTokensQuery>;
};

export default function DeviceEnrollPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(deviceEnrollmentTokensQuery, props.queryRef);
  const [createToken, isCreating] = useCreateDeviceEnrollmentTokenMutation();
  const [secret, setSecret] = useState<string | null>(null);
  const [tokenName, setTokenName] = useState("");

  usePageTitle(__("Enroll device"));

  if (!data.node) return null;

  const tokens = data.node.deviceEnrollmentTokens?.edges.map(e => e.node) ?? [];
  const canCreate = data.node.canCreateEnrollmentToken ?? false;

  const onCreate = () => {
    if (!tokenName.trim()) return;
    void promisifyMutation(createToken)({
      variables: {
        input: {
          organizationId,
          name: tokenName.trim(),
          validitySeconds: 60 * 60 * 24 * 7,
          maxUses: 25,
        },
      },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    }).then((resp: any) => {
      const sec = resp?.createDeviceEnrollmentToken?.secret;
      if (sec) {
        setSecret(sec);
        setTokenName("");
      }
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Enroll device")}
        description={__(
          "Generate an enrollment token, then run the probo-agent install command on the device. The token is single-use-per-device and expires after 7 days.",
        )}
      >
        {canCreate && (
          <div className="flex gap-2">
            <Input
              placeholder={__("Token name (e.g. \"Engineering laptops\")")}
              value={tokenName}
              onChange={e => setTokenName(e.target.value)}
              className="w-72"
            />
            <Button
              onClick={onCreate}
              disabled={isCreating || !tokenName.trim()}
            >
              {__("Generate enrollment token")}
            </Button>
          </div>
        )}
      </PageHeader>

      {secret && (
        <section className="border border-success-border bg-success-bg p-4 rounded">
          <h2 className="font-medium mb-2">
            {__("Enrollment token generated")}
          </h2>
          <p className="text-sm text-tertiary mb-2">
            {__(
              "Copy this token now. It is shown only once — even Probo cannot recover it.",
            )}
          </p>
          <pre className="text-xs bg-surface-default p-3 rounded break-all">
            {secret}
          </pre>
          <h3 className="mt-4 mb-1 text-sm font-medium">
            {__("Install command (macOS / Linux)")}
          </h3>
          <pre className="text-xs bg-surface-default p-3 rounded">
            {`curl -fsSL https://app.getprobo.com/agent/install.sh | sudo PROBO_TOKEN=${secret} sh`}
          </pre>
          <h3 className="mt-4 mb-1 text-sm font-medium">
            {__("Install command (Windows PowerShell)")}
          </h3>
          <pre className="text-xs bg-surface-default p-3 rounded">
            {`$env:PROBO_TOKEN="${secret}"; iwr https://app.getprobo.com/agent/install.ps1 | iex`}
          </pre>
        </section>
      )}

      <section>
        <h2 className="text-lg font-medium mb-2">{__("Active tokens")}</h2>
        <table className="w-full text-sm">
          <Thead>
            <Tr>
              <Th>{__("Name")}</Th>
              <Th>{__("Created at")}</Th>
              <Th>{__("Expires at")}</Th>
              <Th>{__("Usage")}</Th>
              <Th></Th>
            </Tr>
          </Thead>
          <Tbody>
            {tokens.map(t => (
              <TokenRow key={t.id} token={t} />
            ))}
          </Tbody>
        </table>
      </section>
    </div>
  );
}

function TokenRow({
  token,
}: {
  token: {
    id: string;
    name: string;
    createdAt: string;
    expiresAt: string;
    revokedAt?: string | null;
    maxUses?: number | null;
    usedCount: number;
  };
}) {
  const { __ } = useTranslate();
  const revoke = useRevokeDeviceEnrollmentToken(token);
  const revoked = Boolean(token.revokedAt);

  return (
    <Tr>
      <Td>{token.name}</Td>
      <Td>{formatDate(token.createdAt)}</Td>
      <Td>{formatDate(token.expiresAt)}</Td>
      <Td>
        {token.usedCount} / {token.maxUses ?? "∞"}
      </Td>
      <Td className="text-end">
        {!revoked && (
          <Button variant="secondary" onClick={revoke}>
            {__("Revoke")}
          </Button>
        )}
        {revoked && (
          <span className="text-tertiary text-xs">{__("Revoked")}</span>
        )}
      </Td>
    </Tr>
  );
}
