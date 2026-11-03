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
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ScanCounts_scan$key } from "#/__generated__/core/ScanCounts_scan.graphql";

import { scanCounts } from "../../../variants";
import { StatCard } from "../../_components/StatCard";

const scanCountsFragment = graphql`
  fragment ScanCounts_scan on CookieBannerScan {
    pagesScanned
    cookiesFound
    storageFound
    resourcesFound
  }
`;

interface ScanCountsProps {
  scanKey: ScanCounts_scan$key;
}

export function ScanCounts({ scanKey }: ScanCountsProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const scan = useFragment(scanCountsFragment, scanKey);
  const { root, grid } = scanCounts();
  const format = new Intl.NumberFormat(i18n.language);
  const counts = [
    {
      title: t("scanPage.counts.pages"),
      count: scan.pagesScanned,
      icon: <BrowserIcon size={24} aria-hidden />,
    },
    {
      title: t("scanPage.counts.cookies"),
      count: scan.cookiesFound,
      icon: <CookieIcon size={24} aria-hidden />,
    },
    {
      title: t("scanPage.counts.storage"),
      count: scan.storageFound,
      icon: <DatabaseIcon size={24} aria-hidden />,
    },
    {
      title: t("scanPage.counts.resources"),
      count: scan.resourcesFound,
      icon: <CodeIcon size={24} aria-hidden />,
    },
  ];

  return (
    <div className={root()}>
      <div className={grid()}>
        {counts.map(item => (
          <StatCard
            key={item.title}
            icon={item.icon}
            title={item.title}
            value={format.format(item.count)}
          />
        ))}
      </div>
    </div>
  );
}
