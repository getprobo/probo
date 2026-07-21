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

import type { CompliancePageFrameworksSectionFragment$key } from "#/__generated__/core/CompliancePageFrameworksSectionFragment.graphql";

import { CompliancePageFrameworkList } from "./CompliancePageFrameworkList";

const fragment = graphql`
  fragment CompliancePageFrameworksSectionFragment on Organization {
    compliancePage: compliancePortal {
      ...CompliancePageFrameworkList_compliancePageFragment
    }
  }
`;

export function CompliancePageFrameworksSection(props: {
  fragmentRef: CompliancePageFrameworksSectionFragment$key;
}) {
  const { fragmentRef } = props;

  const { t } = useTranslation("organizations/compliance-page");

  const organization = useFragment(fragment, fragmentRef);

  if (!organization.compliancePage) {
    return null;
  }

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-base font-medium">{t("frameworkList.title")}</h2>
        <p className="text-sm text-txt-tertiary">
          {t("frameworkList.description")}
        </p>
      </div>

      <CompliancePageFrameworkList compliancePageRef={organization.compliancePage} />
    </section>
  );
}
