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

import {
  formatAnonymizedIp,
  getActionLabel,
  getActionVariant,
} from "./consentRecordHelpers";

const consentRecordFragment = graphql`
  fragment ConsentRecordRowFragment on CookieConsentRecord {
    id
    visitorId
    action
    cookieBannerVersion {
      version
    }
    ipAddress
    sdkVersion
    createdAt
  }
`;

interface ConsentRecordRowProps {
  recordKey: ConsentRecordRowFragment$key;
}

export function ConsentRecordRow({ recordKey }: ConsentRecordRowProps) {
  const { __ } = useTranslate();
  const record = useFragment(consentRecordFragment, recordKey);

  return (
    <Tr to={record.id}>
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
        <span className="font-mono text-sm">
          {record.ipAddress ? formatAnonymizedIp(record.ipAddress) : "-"}
        </span>
      </Td>
      <Td>
        <span className="font-mono text-sm">{record.sdkVersion}</span>
      </Td>
      <Td>
        <time dateTime={record.createdAt}>
          {formatDate(record.createdAt)}
        </time>
      </Td>
    </Tr>
  );
}
