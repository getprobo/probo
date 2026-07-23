// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageCustomLinksSection_compliancePageFragment$key } from "#/__generated__/core/CompliancePageCustomLinksSection_compliancePageFragment.graphql";

import { CompliancePageCustomLinkList } from "./CompliancePageCustomLinkList";

const compliancePageFragment = graphql`
  fragment CompliancePageCustomLinksSection_compliancePageFragment on CompliancePortal {
    ...CompliancePageCustomLinkList_compliancePageFragment
  }
`;

export interface CompliancePageCustomLinksSectionProps {
  compliancePageRef: CompliancePageCustomLinksSection_compliancePageFragment$key;
}

export function CompliancePageCustomLinksSection(props: CompliancePageCustomLinksSectionProps) {
  const { t } = useTranslation("organizations/compliance-page");

  const compliancePage = useFragment(compliancePageFragment, props.compliancePageRef);

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-base font-medium">{t("externalUrls.title")}</h2>
        <p className="text-sm text-txt-tertiary">
          {t("externalUrls.description")}
        </p>
      </div>

      <CompliancePageCustomLinkList compliancePageRef={compliancePage} />
    </section>
  );
}
