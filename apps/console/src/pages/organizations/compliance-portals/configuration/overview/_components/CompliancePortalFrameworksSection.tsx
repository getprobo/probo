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

import type { CompliancePortalFrameworksSectionFragment$key } from "#/__generated__/core/CompliancePortalFrameworksSectionFragment.graphql";

import { CompliancePortalFrameworkList } from "./CompliancePortalFrameworkList";

const fragment = graphql`
  fragment CompliancePortalFrameworksSectionFragment on CompliancePortal {
    ...CompliancePortalFrameworkList_compliancePortalFragment
  }
`;

export function CompliancePortalFrameworksSection(props: {
  fragmentRef: CompliancePortalFrameworksSectionFragment$key;
}) {
  const { fragmentRef } = props;

  const { t } = useTranslation("organizations/compliance-portals");

  const compliancePortal = useFragment(fragment, fragmentRef);

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-base font-medium">{t("frameworkList.title")}</h2>
        <p className="text-sm text-txt-tertiary">
          {t("frameworkList.description")}
        </p>
      </div>

      <CompliancePortalFrameworkList compliancePortalRef={compliancePortal} />
    </section>
  );
}
