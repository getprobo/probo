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

import { formatDate, formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import {
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceDetailPageQuery } from "#/__generated__/core/DeviceDetailPageQuery.graphql";
import type { DeviceDetailPageRevokeMutation } from "#/__generated__/core/DeviceDetailPageRevokeMutation.graphql";

export const deviceDetailPageQuery = graphql`
  query DeviceDetailPageQuery($deviceId: ID!) {
    device: node(id: $deviceId) @required(action: THROW) {
      __typename
      ... on Device {
        id
        hostname
        hardwareUuid
        serialNumber
        platform
        osVersion
        agentVersion
        enrolledAt
        lastSeenAt
        revokedAt
        latestPostures {
          id
          checkKey
          status
          observedAt
        }
      }
    }
  }
`;

const revokeDeviceMutation = graphql`
  mutation DeviceDetailPageRevokeMutation($input: RevokeDeviceInput!) {
    revokeDevice(input: $input) {
      device {
        id
        revokedAt
      }
    }
  }
`;

interface DeviceDetailPageProps {
  queryRef: PreloadedQuery<DeviceDetailPageQuery>;
}

export function DeviceDetailPage({ queryRef }: DeviceDetailPageProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();

  const { device } = usePreloadedQuery(deviceDetailPageQuery, queryRef);
  if (device.__typename !== "Device") {
    throw new Error("invalid type for device node");
  }

  usePageTitle(device.hostname);

  const [revokeDevice, isRevoking] = useMutation<DeviceDetailPageRevokeMutation>(
    revokeDeviceMutation,
  );

  const isRevoked = Boolean(device.revokedAt);

  const handleRevoke = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          revokeDevice({
            variables: { input: { deviceId: device.id } },
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
                  description: __("Device revoked"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to revoke device"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "Revoke device \"%s\"? The agent on the device will stop reporting and must be re-enrolled.",
          ),
          device.hostname,
        ),
        variant: "danger",
        label: __("Revoke"),
      },
    );
  };

  return (
    <div className="space-y-6">
      <PageHeader title={device.hostname} description={device.platform}>
        {!isRevoked && (
          <Button variant="danger" onClick={handleRevoke} disabled={isRevoking}>
            {__("Revoke")}
          </Button>
        )}
      </PageHeader>

      <section className="grid grid-cols-2 gap-4 max-w-2xl">
        <DetailRow label={__("Hardware UUID")} value={device.hardwareUuid} />
        <DetailRow
          label={__("Serial number")}
          value={device.serialNumber ?? ""}
        />
        <DetailRow label={__("Platform")} value={device.platform} />
        <DetailRow label={__("OS version")} value={device.osVersion} />
        <DetailRow label={__("Agent version")} value={device.agentVersion} />
        <DetailRow
          label={__("Enrolled at")}
          value={formatDate(device.enrolledAt)}
        />
        <DetailRow
          label={__("Last seen")}
          value={device.lastSeenAt ? formatDate(device.lastSeenAt) : __("Never")}
        />
        <DetailRow
          label={__("Revoked")}
          value={device.revokedAt ? formatDate(device.revokedAt) : __("No")}
        />
      </section>

      <section>
        <h2 className="text-lg font-medium mb-2">
          {__("Latest posture checks")}
        </h2>
        <table className="w-full text-sm">
          <Thead>
            <Tr>
              <Th>{__("Check")}</Th>
              <Th>{__("Status")}</Th>
              <Th>{__("Observed at")}</Th>
            </Tr>
          </Thead>
          <Tbody>
            {device.latestPostures.map(p => (
              <Tr key={p.id}>
                <Td>{p.checkKey}</Td>
                <Td>
                  <Badge variant={statusVariant(p.status)}>{p.status}</Badge>
                </Td>
                <Td>{formatDate(p.observedAt)}</Td>
              </Tr>
            ))}
          </Tbody>
        </table>
      </section>
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col">
      <span className="text-tertiary text-xs uppercase">{label}</span>
      <span className="text-sm">{value || "—"}</span>
    </div>
  );
}

function statusVariant(
  status: string,
): "success" | "danger" | "warning" | "info" {
  switch (status) {
    case "PASS":
      return "success";
    case "FAIL":
      return "danger";
    case "NOT_APPLICABLE":
      return "info";
    default:
      return "warning";
  }
}
