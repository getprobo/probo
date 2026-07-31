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

import { graphql, useFragment } from "react-relay";

import type { CompliancePortalHeroMeta_compliancePortal$key } from "./__generated__/CompliancePortalHeroMeta_compliancePortal.graphql";
import { CompliancePortalContactInfo } from "./CompliancePortalContactInfo";
import { CompliancePortalCustomLinks } from "./CompliancePortalCustomLinks";
import { organizationContactInfo } from "./variants";

const compliancePortalHeroMetaFragment = graphql`
  fragment CompliancePortalHeroMeta_compliancePortal on CompliancePortal {
    websiteUrl
    email
    headquarterAddress
    customLinks(first: 20) {
      edges {
        __typename
      }
    }
    ...CompliancePortalContactInfo_compliancePortal
    ...CompliancePortalCustomLinks_compliancePortal
  }
`;

interface CompliancePortalHeroMetaProps {
  compliancePortalKey: CompliancePortalHeroMeta_compliancePortal$key;
}

// Shared hero bottom band: contact details and custom links in one row with a
// single top divider. Hidden entirely when both are empty.
export function CompliancePortalHeroMeta({
  compliancePortalKey,
}: CompliancePortalHeroMetaProps) {
  const compliancePortal = useFragment(
    compliancePortalHeroMetaFragment,
    compliancePortalKey,
  );

  const hasWebsite = compliancePortal.websiteUrl != null && compliancePortal.websiteUrl !== "";
  const hasEmail = compliancePortal.email != null && compliancePortal.email !== "";
  const hasAddress
    = compliancePortal.headquarterAddress != null && compliancePortal.headquarterAddress !== "";
  const hasContact = hasWebsite || hasEmail || hasAddress;
  const hasCustomLinks = compliancePortal.customLinks.edges.length > 0;

  if (!hasContact && !hasCustomLinks) {
    return null;
  }

  const { root } = organizationContactInfo();

  return (
    <div className={root()}>
      <CompliancePortalContactInfo compliancePortalKey={compliancePortal} />
      <CompliancePortalCustomLinks compliancePortalKey={compliancePortal} />
    </div>
  );
}
