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
    regulation
    regulationSource
    countryCode
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
        <span className="font-mono text-sm">
          {record.regulation || "-"}
        </span>
      </Td>
      <Td>
        <span className="font-mono text-sm">
          {record.regulationSource || "-"}
        </span>
      </Td>
      <Td>
        <span className="font-mono text-sm">
          {record.countryCode || "-"}
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
