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

import { ArrowSquareOutIcon } from "@phosphor-icons/react";
import { externalLinkProps } from "@probo/helpers";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { CompliancePortalOpenLink_compliancePortal$key } from "#/__generated__/core/CompliancePortalOpenLink_compliancePortal.graphql";
import { navPanelSwitcher } from "#/pages/organizations/_components/NavPanelSwitcher";

const compliancePortalOpenLinkFragment = graphql`
  fragment CompliancePortalOpenLink_compliancePortal on CompliancePortal {
    active
    publicUrl
  }
`;

export interface CompliancePortalOpenLinkProps {
  compliancePortalKey: CompliancePortalOpenLink_compliancePortal$key;
}

/**
 * Icon that opens the selected portal in a new tab.
 *
 * Hidden while that portal is inactive or has no public URL — the same gate
 * as the page-header Open button.
 */
export function CompliancePortalOpenLink({
  compliancePortalKey,
}: CompliancePortalOpenLinkProps) {
  const { t } = useTranslation();
  const { active, publicUrl } = useFragment(
    compliancePortalOpenLinkFragment,
    compliancePortalKey,
  );
  if (!active || publicUrl === "") {
    return null;
  }
  const link = externalLinkProps(publicUrl);
  if (link.href == null) {
    return null;
  }

  const slots = navPanelSwitcher();

  return (
    <a
      {...link}
      className={slots.openLink()}
      aria-label={t("nav.compliancePortalSwitcher.open")}
    >
      <ArrowSquareOutIcon size={16} aria-hidden />
    </a>
  );
}
