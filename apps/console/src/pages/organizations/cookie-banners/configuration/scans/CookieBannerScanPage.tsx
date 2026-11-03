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

import { useEffect, useTransition } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery, useRefetchableFragment } from "react-relay";

import type { CookieBannerScanPage_scan$key } from "#/__generated__/core/CookieBannerScanPage_scan.graphql";
import type { CookieBannerScanPageQuery } from "#/__generated__/core/CookieBannerScanPageQuery.graphql";
import type { CookieBannerScanPageRefetchQuery } from "#/__generated__/core/CookieBannerScanPageRefetchQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { ScanCheckpoint } from "./_components/ScanCheckpoint";
import { ScanCounts } from "./_components/ScanCounts";
import { ScanDetailHeader } from "./_components/ScanDetailHeader";
import { ScanTimeline } from "./_components/ScanTimeline";
import { SCAN_POLL_MS } from "./_lib/scanPoll";
import { isActiveScanState } from "./_lib/scanStates";

export const cookieBannerScanPageQuery = graphql`
  query CookieBannerScanPageQuery($scanId: ID!) {
    node(id: $scanId) @required(action: THROW) {
      __typename
      ... on CookieBannerScan {
        ...CookieBannerScanPage_scan
      }
    }
  }
`;

const cookieBannerScanPageFragment = graphql`
  fragment CookieBannerScanPage_scan on CookieBannerScan
  @refetchable(queryName: "CookieBannerScanPageRefetchQuery") {
    state
    ...ScanDetailHeader_scan
    ...ScanCounts_scan
    ...ScanTimeline_scan
    ...ScanCheckpoint_scan
  }
`;

interface CookieBannerScanPageProps {
  queryRef: PreloadedQuery<CookieBannerScanPageQuery>;
}

export default function CookieBannerScanPage({ queryRef }: CookieBannerScanPageProps) {
  const data = usePreloadedQuery<CookieBannerScanPageQuery>(cookieBannerScanPageQuery, queryRef);

  if (data.node.__typename !== "CookieBannerScan") {
    throw new NotFoundError("Cookie banner scan not found");
  }

  return <CookieBannerScanPageContent scanKey={data.node} />;
}

function CookieBannerScanPageContent({ scanKey }: { scanKey: CookieBannerScanPage_scan$key }) {
  const [, startRefetchTransition] = useTransition();
  const [scan, refetch] = useRefetchableFragment<
    CookieBannerScanPageRefetchQuery,
    CookieBannerScanPage_scan$key
  >(cookieBannerScanPageFragment, scanKey);
  const active = isActiveScanState(scan.state);

  useEffect(() => {
    if (!active) {
      return;
    }
    const interval = setInterval(() => {
      if (document.hidden) {
        return;
      }
      startRefetchTransition(() => {
        refetch({}, { fetchPolicy: "network-only" });
      });
    }, SCAN_POLL_MS);
    return () => clearInterval(interval);
  }, [active, refetch]);

  return (
    <div className="space-y-6">
      <ScanDetailHeader scanKey={scan} />
      <ScanTimeline scanKey={scan} />
      <ScanCounts scanKey={scan} />
      <ScanCheckpoint scanKey={scan} />
    </div>
  );
}
