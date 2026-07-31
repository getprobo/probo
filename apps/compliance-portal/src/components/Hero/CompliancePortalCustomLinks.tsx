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

import {
  FacebookLogoIcon,
  GlobeSimpleIcon,
  type Icon,
  LinkedinLogoIcon,
  XLogoIcon,
} from "@phosphor-icons/react";
import { detectSocialName } from "@probo/helpers";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import { graphql, useFragment } from "react-relay";

import { externalHref } from "#/lib/url/hostname";

import type { CompliancePortalCustomLinks_compliancePortal$key } from "./__generated__/CompliancePortalCustomLinks_compliancePortal.graphql";
import { organizationContactInfo } from "./variants";

const compliancePortalCustomLinksFragment = graphql`
  fragment CompliancePortalCustomLinks_compliancePortal on CompliancePortal {
    customLinks(first: 20) {
      edges {
        node {
          id
          name
          url
        }
      }
    }
  }
`;

interface CompliancePortalCustomLinksProps {
  compliancePortalKey: CompliancePortalCustomLinks_compliancePortal$key;
}

function iconForUrl(url: string): Icon {
  switch (detectSocialName(url)) {
    case "LinkedIn":
      return LinkedinLogoIcon;
    case "X":
      return XLogoIcon;
    case "Facebook":
      return FacebookLogoIcon;
    default:
      return GlobeSimpleIcon;
  }
}

// Organization custom links (social / external URLs) as icon + label items,
// appended after contact details in the shared hero meta row.
export function CompliancePortalCustomLinks({
  compliancePortalKey,
}: CompliancePortalCustomLinksProps) {
  const compliancePortal = useFragment(
    compliancePortalCustomLinksFragment,
    compliancePortalKey,
  );
  const links = compliancePortal.customLinks.edges.map(edge => edge.node);

  if (links.length === 0) {
    return null;
  }

  const { link } = organizationContactInfo();

  return (
    <>
      {links.map((customLink) => {
        const Icon = iconForUrl(customLink.url);

        return (
          <Anchor
            key={customLink.id}
            className={link()}
            href={externalHref(customLink.url)}
            target="_blank"
            rel="noopener noreferrer"
            size={2}
            color="neutral"
            underline={false}
            iconStart={<Icon />}
          >
            {customLink.name}
          </Anchor>
        );
      })}
    </>
  );
}
