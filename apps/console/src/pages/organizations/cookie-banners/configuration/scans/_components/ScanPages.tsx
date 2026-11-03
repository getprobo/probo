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

import { CheckIcon } from "@phosphor-icons/react";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ScanPages_scan$key } from "#/__generated__/core/ScanPages_scan.graphql";

import { cookieBannerList } from "../../../variants";
import { SCAN_FINDINGS_PAGE_SIZE, SCAN_MAX_PAGES } from "../_lib/pageSize";
import { type ScanPageRow, scanPageRows } from "../_lib/scanPages";

import { FindingsPagination } from "./FindingsPagination";

const pageFilters = ["scanned", "notScanned"] as const;
type PageFilter = (typeof pageFilters)[number];

function isPageFilter(value: string): value is PageFilter {
  return (pageFilters as readonly string[]).includes(value);
}

const scanPagesFragment = graphql`
  fragment ScanPages_scan on CookieBannerScan {
    truncated
    checkpoint {
      visitedTemplates
      queue {
        url
      }
    }
  }
`;

interface ScanPagesProps {
  scanKey: ScanPages_scan$key;
}

export function ScanPages({ scanKey }: ScanPagesProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const scan = useFragment(scanPagesFragment, scanKey);
  const [filter, setFilter] = useState<PageFilter | null>(null);
  const [page, setPage] = useState(0);
  const { filter: filterSlot } = cookieBannerList();
  const allLabel = t("scanPage.findings.pagesFilter.all");
  const rows = scanPageRows(
    scan.checkpoint?.visitedTemplates ?? [],
    scan.checkpoint?.queue ?? [],
  );
  const filtered = rows.filter((row) => {
    if (filter === "scanned") {
      return row.scanned;
    }
    if (filter === "notScanned") {
      return !row.scanned;
    }
    return true;
  });
  const pageCount = Math.max(1, Math.ceil(filtered.length / SCAN_FINDINGS_PAGE_SIZE));
  const pageIndex = Math.min(page, pageCount - 1);
  const visible = filtered.slice(
    pageIndex * SCAN_FINDINGS_PAGE_SIZE,
    (pageIndex + 1) * SCAN_FINDINGS_PAGE_SIZE,
  );

  function selectFilter(next: PageFilter | null) {
    setFilter(next);
    setPage(0);
  }

  let emptyKey = "scanPage.findings.pagesEmpty";
  if (scan.checkpoint == null) {
    emptyKey = "scanPage.findings.pagesNoCheckpoint";
  } else if (filter != null) {
    emptyKey = "scanPage.findings.pagesFilteredEmpty";
  }

  return (
    <div className="flex flex-col gap-3">
      {scan.truncated
        ? (
            <Callout color="neutral">
              {t("scanPage.findings.truncated", { count: SCAN_MAX_PAGES })}
            </Callout>
          )
        : null}
      <div className={filterSlot()}>
        <Select
          value={filter}
          onValueChange={(value: string | null) => {
            if (value == null) {
              selectFilter(null);
              return;
            }
            if (isPageFilter(value)) {
              selectFilter(value);
            }
          }}
        >
          <SelectTrigger
            size={2}
            placeholder={allLabel}
            aria-label={t("scanPage.findings.pagesFilter.ariaLabel")}
          >
            {(value: PageFilter | null) => (
              value != null ? t(`scanPage.findings.pagesFilter.${value}`) : allLabel
            )}
          </SelectTrigger>
          <SelectPopup align="start">
            <SelectItem value={null}>{allLabel}</SelectItem>
            {pageFilters.map(pageFilter => (
              <SelectItem key={pageFilter} value={pageFilter}>
                {t(`scanPage.findings.pagesFilter.${pageFilter}`)}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
      {visible.length === 0
        ? <Text size={2} color="faint">{t(emptyKey)}</Text>
        : (
            <>
              <Table variant="surface" layout="fixed">
                <TableHeader>
                  <TableRow>
                    <TableColumnHeaderCell overflow="truncate">
                      {t("scanPage.findings.page")}
                    </TableColumnHeaderCell>
                    <TableColumnHeaderCell width="7rem">
                      {t("scanPage.findings.scanned")}
                    </TableColumnHeaderCell>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {visible.map(row => (
                    <PageRow key={row.url} row={row} scannedLabel={t("scanPage.findings.scanned")} />
                  ))}
                </TableBody>
              </Table>
              <FindingsPagination
                hasPrevious={pageIndex > 0}
                hasNext={pageIndex < pageCount - 1}
                onPrevious={() => setPage(Math.max(0, pageIndex - 1))}
                onNext={() => setPage(Math.min(pageCount - 1, pageIndex + 1))}
              />
            </>
          )}
    </div>
  );
}

function PageRow({ row, scannedLabel }: { row: ScanPageRow; scannedLabel: string }) {
  return (
    <TableRow>
      <TableCell overflow="truncate">
        <Text size={2} className="break-all font-mono">{row.url}</Text>
      </TableCell>
      <TableCell>
        {row.scanned
          ? <CheckIcon aria-label={scannedLabel} />
          : null}
      </TableCell>
    </TableRow>
  );
}
