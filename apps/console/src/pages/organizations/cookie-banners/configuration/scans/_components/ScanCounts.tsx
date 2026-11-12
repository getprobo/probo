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

import { BrowserIcon, CodeIcon, CookieIcon, DatabaseIcon } from "@phosphor-icons/react";
import { SegmentedControl } from "@probo/ui/src/v2/SegmentedControl/SegmentedControl";
import { SegmentedControlItem } from "@probo/ui/src/v2/SegmentedControl/SegmentedControlItem";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ScanCounts_scan$key } from "#/__generated__/core/ScanCounts_scan.graphql";

import { cookieBannerList } from "../../../variants";
import { scanPageRows, scanPageTotals } from "../_lib/scanPages";

import { ScanCookies, ScanStorage } from "./ScanDetectedTrackers";
import { ScanPages } from "./ScanPages";
import { ScanResources } from "./ScanResources";

const scanCountsFragment = graphql`
  fragment ScanCounts_scan on CookieBannerScan {
    checkpoint {
      pagesScanned
      visitedTemplates
      queue {
        url
      }
    }
    cookiesInfo: detectedTrackers(first: 0, filter: { trackerTypes: [COOKIE] }) {
      totalCount
    }
    storageInfo: detectedTrackers(
      first: 0
      filter: { trackerTypes: [LOCAL_STORAGE, SESSION_STORAGE] }
    ) {
      totalCount
    }
    resourcesInfo: trackerResources(first: 0) {
      totalCount
    }
    ...ScanPages_scan
    ...ScanDetectedTrackers_cookies
    ...ScanDetectedTrackers_storage
    ...ScanResources_scan
  }
`;

const scanTabs = ["pages", "cookies", "storage", "resources"] as const;
type ScanTab = (typeof scanTabs)[number];

function isScanTab(value: string): value is ScanTab {
  return (scanTabs as readonly string[]).includes(value);
}

interface ScanCountsProps {
  scanKey: ScanCounts_scan$key;
}

export function ScanCounts({ scanKey }: ScanCountsProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const scan = useFragment(scanCountsFragment, scanKey);
  const [tab, setTab] = useState<ScanTab>("pages");
  const { views } = cookieBannerList();
  const format = new Intl.NumberFormat(i18n.language);
  const countClass = "font-mono tabular-nums";
  const pageRows = scanPageRows(
    scan.checkpoint?.visitedTemplates ?? [],
    (scan.checkpoint?.queue ?? []).map(item => ({ url: item.url })),
  );
  const pageTotals = scanPageTotals(pageRows, scan.checkpoint?.pagesScanned ?? 0);

  return (
    <div className="flex flex-col gap-4">
      <div className={views()}>
        <SegmentedControl
          size={1}
          value={tab}
          aria-label={t("scanPage.counts.ariaLabel")}
          onValueChange={(value) => {
            if (isScanTab(value)) {
              setTab(value);
            }
          }}
        >
          <SegmentedControlItem value="pages">
            <BrowserIcon />
            {t("scanPage.counts.pages")}
            <span className={countClass}>
              {format.format(pageTotals.scanned)}
              {" / "}
              {format.format(pageTotals.total)}
            </span>
          </SegmentedControlItem>
          <SegmentedControlItem value="cookies">
            <CookieIcon />
            {t("scanPage.counts.cookies")}
            <span className={countClass}>{format.format(scan.cookiesInfo.totalCount)}</span>
          </SegmentedControlItem>
          <SegmentedControlItem value="storage">
            <DatabaseIcon />
            {t("scanPage.counts.storage")}
            <span className={countClass}>{format.format(scan.storageInfo.totalCount)}</span>
          </SegmentedControlItem>
          <SegmentedControlItem value="resources">
            <CodeIcon />
            {t("scanPage.counts.resources")}
            <span className={countClass}>{format.format(scan.resourcesInfo.totalCount)}</span>
          </SegmentedControlItem>
        </SegmentedControl>
      </div>
      {tab === "pages" ? <ScanPages scanKey={scan} /> : null}
      {tab === "cookies" ? <ScanCookies scanKey={scan} /> : null}
      {tab === "storage" ? <ScanStorage scanKey={scan} /> : null}
      {tab === "resources" ? <ScanResources scanKey={scan} /> : null}
    </div>
  );
}
