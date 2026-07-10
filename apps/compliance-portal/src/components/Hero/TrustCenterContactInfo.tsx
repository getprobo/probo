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

import { EnvelopeIcon, GlobeSimpleIcon, MapPinSimpleIcon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { externalHref, hostnameOf } from "#/lib/url/hostname";

import type { TrustCenterContactInfo_trustCenter$key } from "./__generated__/TrustCenterContactInfo_trustCenter.graphql";
import { organizationContactInfo } from "./variants";

const trustCenterContactInfoFragment = graphql`
  fragment TrustCenterContactInfo_trustCenter on TrustCenter {
    websiteUrl
    email
    headquarterAddress
  }
`;

interface TrustCenterContactInfoProps {
  trustCenterKey: TrustCenterContactInfo_trustCenter$key;
}

// Trust center contact details (website, email, HQ) rendered as an icon + label
// row. Owns its fragment so it can be reused wherever the trust center is in scope.
export function TrustCenterContactInfo({ trustCenterKey }: TrustCenterContactInfoProps) {
  const trustCenter = useFragment(trustCenterContactInfoFragment, trustCenterKey);

  const hasWebsite = trustCenter.websiteUrl != null && trustCenter.websiteUrl !== "";
  const hasEmail = trustCenter.email != null && trustCenter.email !== "";
  const hasAddress = trustCenter.headquarterAddress != null && trustCenter.headquarterAddress !== "";

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
