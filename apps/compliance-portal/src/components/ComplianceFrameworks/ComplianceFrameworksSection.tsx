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
import { graphql, useFragment } from "react-relay";

import { HomeSection } from "#/components/HomeSection/HomeSection";

import type { ComplianceFrameworksSection_trustCenter$key } from "./__generated__/ComplianceFrameworksSection_trustCenter.graphql";
import { ComplianceFrameworkListItem } from "./ComplianceFrameworkListItem";

const complianceFrameworksSectionFragment = graphql`
  fragment ComplianceFrameworksSection_trustCenter on TrustCenter {
    complianceFrameworks(first: 8) {
      edges {
        node {
          id
          ...ComplianceFrameworkListItem_complianceFramework
        }
      }
    }
  }
`;

interface ComplianceFrameworksSectionProps {
  trustCenterKey: ComplianceFrameworksSection_trustCenter$key;
}

// "Compliance" section: the grid of certification frameworks the trust center
// covers.
export function ComplianceFrameworksSection({ trustCenterKey }: ComplianceFrameworksSectionProps) {
  const { t } = useTranslation();
  const data = useFragment(complianceFrameworksSectionFragment, trustCenterKey);
  const frameworks = data.complianceFrameworks.edges.map(edge => edge.node);

  if (frameworks.length === 0) {
    return null;
  }

  return (
    <HomeSection title={t("home.sections.compliance")}>
      <div className="grid grid-cols-6 gap-4">
        {frameworks.map(framework => (
          <ComplianceFrameworkListItem key={framework.id} complianceFrameworkKey={framework} />
        ))}
      </div>
    </HomeSection>
  );
}
