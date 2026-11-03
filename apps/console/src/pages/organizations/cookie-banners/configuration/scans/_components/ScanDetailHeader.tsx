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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { ScanDetailHeader_scan$key } from "#/__generated__/core/ScanDetailHeader_scan.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { cookieBannerScansPath } from "../../../_lib/cookieBannerPaths";
import { trackerPatternDetailHeader } from "../../../variants";
import { formatScanTime } from "../_lib/formatScanTime";

import { ScanStateBadge } from "./ScanStateBadge";

const scanDetailHeaderFragment = graphql`
  fragment ScanDetailHeader_scan on CookieBannerScan {
    state
    createdAt
  }
`;

interface ScanDetailHeaderProps {
  scanKey: ScanDetailHeader_scan$key;
}

export function ScanDetailHeader({ scanKey }: ScanDetailHeaderProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const location = useLocation();
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const scan = useFragment(scanDetailHeaderFragment, scanKey);
  const requestedLabel = formatScanTime(i18n.language, scan.createdAt);
  usePageTitle(requestedLabel);
  const { root, back, bar, titleRow, title, badges } = trackerPatternDetailHeader();

  if (cookieBannerId == null) {
    throw new Error(":cookieBannerId missing in route params");
  }

  return (
    <div className={root()}>
      <Link
        to={{ pathname: cookieBannerScansPath(organizationId, cookieBannerId), search: location.search }}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("scanPage.actions.back")}
      </Link>
      <div className={bar()}>
        <div className={titleRow()}>
          <Heading level={1} size={6} weight="medium" highContrast className={title()}>
            {requestedLabel}
          </Heading>
          <div className={badges()}>
            <ScanStateBadge state={scan.state} />
          </div>
        </div>
      </div>
    </div>
  );
}
