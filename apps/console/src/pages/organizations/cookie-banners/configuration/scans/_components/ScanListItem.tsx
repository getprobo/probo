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

import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableLink } from "@probo/ui/src/v2/Table/TableLink";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { ScanListItem_scan$key } from "#/__generated__/core/ScanListItem_scan.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { cookieBannerScanPath } from "../../../_lib/cookieBannerPaths";
import { trackerPatternListItem } from "../../../variants";
import { formatScanTime } from "../_lib/formatScanTime";

import { ScanStateBadge, ScanTruncatedBadge } from "./ScanStateBadge";

const scanListItemFragment = graphql`
  fragment ScanListItem_scan on CookieBannerScan {
    id
    state
    truncated
    pagesScanned
    cookiesFound
    storageFound
    resourcesFound
    startedAt
    completedAt
    createdAt
  }
`;

interface ScanListItemProps {
  scanKey: ScanListItem_scan$key;
}

export function ScanListItem({ scanKey }: ScanListItemProps) {
  const { i18n } = useTranslation("organizations/cookie-banners");
  const location = useLocation();
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const scan = useFragment(scanListItemFragment, scanKey);
  const { optional, date } = trackerPatternListItem();

  if (cookieBannerId == null) {
    throw new Error(":cookieBannerId missing in route params");
  }

  const requested = formatScanTime(i18n.language, scan.createdAt);

  return (
    <TableRow align="center" interactive>
      <TableRowHeaderCell overflow="truncate">
        <TableLink
          to={{
            pathname: cookieBannerScanPath(organizationId, cookieBannerId, scan.id),
            search: location.search,
          }}
        >
          <Text size={2} weight="medium" highContrast className={date()}>
            {requested}
          </Text>
        </TableLink>
      </TableRowHeaderCell>
      <TableCell>
        <div className="flex flex-wrap items-center gap-1">
          <ScanStateBadge state={scan.state} />
          {scan.truncated ? <ScanTruncatedBadge /> : null}
        </div>
      </TableCell>
      <TableCell className={optional()}>
        <Text size={2} className={date()}>{formatScanTime(i18n.language, scan.startedAt)}</Text>
      </TableCell>
      <TableCell className={optional()}>
        <Text size={2} className={date()}>{formatScanTime(i18n.language, scan.completedAt)}</Text>
      </TableCell>
      <TableCell>
        <Text size={2} className="tabular-nums">{scan.pagesScanned}</Text>
      </TableCell>
      <TableCell>
        <Text size={2} className="tabular-nums">{scan.cookiesFound}</Text>
      </TableCell>
      <TableCell>
        <Text size={2} className="tabular-nums">{scan.storageFound}</Text>
      </TableCell>
      <TableCell>
        <Text size={2} className="tabular-nums">{scan.resourcesFound}</Text>
      </TableCell>
    </TableRow>
  );
}
