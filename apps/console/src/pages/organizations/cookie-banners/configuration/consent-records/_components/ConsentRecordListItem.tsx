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

import { dateFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableLink } from "@probo/ui/src/v2/Table/TableLink";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { ConsentRecordListItem_cookieConsentRecord$key } from "#/__generated__/core/ConsentRecordListItem_cookieConsentRecord.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { cookieBannerConsentRecordPath } from "../../../_lib/cookieBannerPaths";
import {
  formatLocation,
  getActionBadgeColor,
} from "../_lib/consentRecordHelpers";

const consentRecordListItemFragment = graphql`
  fragment ConsentRecordListItem_cookieConsentRecord on CookieConsentRecord {
    id
    action
    cookieBannerVersion {
      version
    }
    regulation
    countryCode
    subdivisionCode
    origin
    createdAt
  }
`;

interface ConsentRecordListItemProps {
  cookieConsentRecordKey: ConsentRecordListItem_cookieConsentRecord$key;
  corsless: boolean;
}

export function ConsentRecordListItem({
  cookieConsentRecordKey,
  corsless,
}: ConsentRecordListItemProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const record = useFragment(consentRecordListItemFragment, cookieConsentRecordKey);
  if (typeof cookieBannerId !== "string") {
    throw new Error("Missing cookieBannerId parameter");
  }
  const location = formatLocation(record.countryCode, record.subdivisionCode);
  const detailPath = cookieBannerConsentRecordPath(
    organizationId,
    cookieBannerId,
    record.id,
  );

  return (
    <TableRow align="center" interactive>
      <TableRowHeaderCell>
        <TableLink to={detailPath}>
          <Badge variant="soft" color={getActionBadgeColor(record.action)}>
            {t(`consentRecordPage.actions.${record.action.toLowerCase()}`)}
          </Badge>
        </TableLink>
      </TableRowHeaderCell>
      <TableCell>
        {record.regulation
          ? (
              <Text size={2} highContrast>
                {record.regulation}
              </Text>
            )
          : (
              <Text size={2} color="faint">—</Text>
            )}
      </TableCell>
      <TableCell>
        {location === ""
          ? <Text size={2} color="faint">—</Text>
          : (
              <Text size={2} highContrast>
                {location}
              </Text>
            )}
      </TableCell>
      <TableCell>
        {record.cookieBannerVersion
          ? (
              <Text size={2} highContrast>
                {t("configLayout.callout.version", { version: record.cookieBannerVersion.version })}
              </Text>
            )
          : <Text size={2} color="faint">—</Text>}
      </TableCell>
      {corsless
        ? (
            <TableCell overflow="truncate">
              {record.origin
                ? (
                    <Text size={2} highContrast>
                      {record.origin}
                    </Text>
                  )
                : <Text size={2} color="faint">—</Text>}
            </TableCell>
          )
        : null}
      <TableCell>
        <Text size={2} highContrast>
          <time dateTime={record.createdAt}>
            {dateFormat(i18n.language, record.createdAt)}
          </time>
        </Text>
      </TableCell>
    </TableRow>
  );
}
