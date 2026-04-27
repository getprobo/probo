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

import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Td, Tr } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { ConsentRecordRowFragment$key } from "#/__generated__/core/ConsentRecordRowFragment.graphql";

const consentRecordFragment = graphql`
  fragment ConsentRecordRowFragment on CookieConsentRecord {
    visitorId
    action
    cookieBannerVersion {
      id
      version
    }
    ipAddress
    sdkVersion
    consentData
    createdAt
  }
`;

function getActionLabel(action: string, __: (s: string) => string): string {
  switch (action) {
    case "ACCEPT_ALL":
      return __("Accept All");
    case "REJECT_ALL":
      return __("Reject All");
    case "CUSTOMIZE":
      return __("Customize");
    case "GPC":
      return __("GPC");
    default:
      return action;
  }
}

function getActionVariant(action: string): "success" | "danger" | "warning" | "neutral" {
  switch (action) {
    case "ACCEPT_ALL":
      return "success";
    case "REJECT_ALL":
      return "danger";
    case "CUSTOMIZE":
      return "warning";
    case "GPC":
      return "neutral";
    default:
      return "neutral";
  }
}

interface ConsentRecordRowProps {
  recordKey: ConsentRecordRowFragment$key;
}

export function ConsentRecordRow({ recordKey }: ConsentRecordRowProps) {
  const { __ } = useTranslate();
  const record = useFragment(consentRecordFragment, recordKey);

  return (
    <Tr>
      <Td>
        <span className="font-mono text-sm">{record.visitorId}</span>
      </Td>
      <Td>
        <Badge variant={getActionVariant(record.action)}>
          {getActionLabel(record.action, __)}
        </Badge>
      </Td>
      <Td>
        {record.cookieBannerVersion
          ? (
              <span className="font-mono text-sm">
                {record.cookieBannerVersion.version}
              </span>
            )
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        <span className="font-mono text-sm">{record.ipAddress ?? "-"}</span>
      </Td>
      <Td>
        <span className="font-mono text-sm">{record.sdkVersion}</span>
      </Td>
      <Td>
        <span className="font-mono text-xs max-w-48 truncate block" title={record.consentData}>
          {record.consentData}
        </span>
      </Td>
      <Td>
        <time dateTime={record.createdAt}>
          {formatDate(record.createdAt)}
        </time>
      </Td>
    </Tr>
  );
}
