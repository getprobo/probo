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

import { EnvelopeIcon, GlobeSimpleIcon, MapPinSimpleIcon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { externalHref, hostnameOf } from "#/lib/url/hostname";

import type { CompliancePageContactInfo_trustCenter$key } from "./__generated__/CompliancePageContactInfo_trustCenter.graphql";
import { compliancePageContactInfo } from "./variants";

const compliancePageContactInfoFragment = graphql`
  fragment CompliancePageContactInfo_trustCenter on TrustCenter {
    websiteUrl
    email
    headquarterAddress
  }
`;

interface CompliancePageContactInfoProps {
  trustCenterKey: CompliancePageContactInfo_trustCenter$key;
}

// Compliance page contact details (website, email, HQ) rendered as an icon + label
// row. Owns its fragment so it can be reused wherever the trust center is in scope.
export function CompliancePageContactInfo({ trustCenterKey }: CompliancePageContactInfoProps) {
  const trustCenter = useFragment(compliancePageContactInfoFragment, trustCenterKey);

  const hasWebsite = trustCenter.websiteUrl != null && trustCenter.websiteUrl !== "";
  const hasEmail = trustCenter.email != null && trustCenter.email !== "";
  const hasAddress = trustCenter.headquarterAddress != null && trustCenter.headquarterAddress !== "";

  // Nothing to show — render no row (and therefore no divider) at all.
  if (!hasWebsite && !hasEmail && !hasAddress) {
    return null;
  }

  const { root, item, link } = compliancePageContactInfo();

  return (
    <div className={root()}>
      {hasWebsite && (
        <a
          className={link()}
          href={externalHref(trustCenter.websiteUrl)}
          target="_blank"
          rel="noopener noreferrer"
        >
          <GlobeSimpleIcon />
          <Text size={2} color="neutral">
            {hostnameOf(trustCenter.websiteUrl)}
          </Text>
        </a>
      )}
      {hasEmail && (
        <a className={link()} href={`mailto:${trustCenter.email}`}>
          <EnvelopeIcon />
          <Text size={2} color="neutral">
            {trustCenter.email}
          </Text>
        </a>
      )}
      {hasAddress && (
        <div className={item()}>
          <MapPinSimpleIcon />
          <Text size={2} color="neutral">
            {trustCenter.headquarterAddress}
          </Text>
        </div>
      )}
    </div>
  );
}
