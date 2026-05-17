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

import { formatDate } from "@probo/helpers";
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
} from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { DeviceGraphNodeQuery } from "#/__generated__/core/DeviceGraphNodeQuery.graphql";
import { deviceNodeQuery, useRevokeDevice } from "#/hooks/graph/DeviceGraph";

type Props = {
  queryRef: PreloadedQuery<DeviceGraphNodeQuery>;
};

export default function DeviceDetailPage(props: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(deviceNodeQuery, props.queryRef);
  const device = data.node;
  const revoke = useRevokeDevice({ id: device?.id, hostname: device?.hostname });

  usePageTitle(device?.hostname ?? __("Device"));

  if (!device) return null;

  return (
    <div className="space-y-6">
      <PageHeader
        title={device.hostname ?? __("Device")}
        description={device.platform ?? ""}
      >
        <Button variant="danger" onClick={revoke}>
          {__("Revoke")}
        </Button>
      </PageHeader>

      <section className="grid grid-cols-2 gap-4 max-w-2xl">
        <DetailRow label={__("Hardware UUID")} value={device.hardwareUuid} />
        <DetailRow label={__("Serial number")} value={device.serialNumber ?? ""} />
        <DetailRow label={__("Platform")} value={device.platform ?? ""} />
        <DetailRow label={__("OS version")} value={device.osVersion ?? ""} />
        <DetailRow label={__("Agent version")} value={device.agentVersion ?? ""} />
        <DetailRow
          label={__("Enrolled at")}
          value={device.enrolledAt ? formatDate(device.enrolledAt) : ""}
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
        <h2 className="text-lg font-medium mb-2">{__("Latest posture checks")}</h2>
        <table className="w-full text-sm">
          <Thead>
            <Tr>
              <Th>{__("Check")}</Th>
              <Th>{__("Status")}</Th>
              <Th>{__("Observed at")}</Th>
            </Tr>
          </Thead>
          <Tbody>
            {device.latestPostures?.map(p => (
              <Tr key={p.checkKey}>
                <Td>{p.checkKey}</Td>
                <Td>
                  <Badge variant={statusVariant(p.status)}>{p.status}</Badge>
                </Td>
                <Td>{p.observedAt ? formatDate(p.observedAt) : ""}</Td>
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
