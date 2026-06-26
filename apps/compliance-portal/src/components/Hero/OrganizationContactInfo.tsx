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

import { hostnameOf } from "#/lib/url/hostname";

import type { OrganizationContactInfo_organization$key } from "./__generated__/OrganizationContactInfo_organization.graphql";
import { organizationContactInfo } from "./variants";

const organizationContactInfoFragment = graphql`
  fragment OrganizationContactInfo_organization on Organization {
    websiteUrl
    email
    headquarterAddress
  }
`;

interface OrganizationContactInfoProps {
  organizationKey: OrganizationContactInfo_organization$key;
}

// Organization contact details (website, email, HQ) rendered as an icon + label
// row. Owns its fragment so it can be reused wherever the org is in scope.
export function OrganizationContactInfo({ organizationKey }: OrganizationContactInfoProps) {
  const organization = useFragment(organizationContactInfoFragment, organizationKey);

  const hasWebsite = organization.websiteUrl != null && organization.websiteUrl !== "";
  const hasEmail = organization.email != null && organization.email !== "";
  const hasAddress = organization.headquarterAddress != null && organization.headquarterAddress !== "";

  // Nothing to show — render no row (and therefore no divider) at all.
  if (!hasWebsite && !hasEmail && !hasAddress) {
    return null;
  }

  const { root, item, link } = organizationContactInfo();

  return (
    <div className={root()}>
      {hasWebsite && (
        <a
          className={link()}
          href={organization.websiteUrl}
          target="_blank"
          rel="noopener noreferrer"
        >
          <GlobeSimpleIcon />
          <Text size={2} color="neutral">
            {hostnameOf(organization.websiteUrl)}
          </Text>
        </a>
      )}
      {hasEmail && (
        <a className={link()} href={`mailto:${organization.email}`}>
          <EnvelopeIcon />
          <Text size={2} color="neutral">
            {organization.email}
          </Text>
        </a>
      )}
      {hasAddress && (
        <div className={item()}>
          <MapPinSimpleIcon />
          <Text size={2} color="neutral">
            {organization.headquarterAddress}
          </Text>
        </div>
      )}
    </div>
  );
}
