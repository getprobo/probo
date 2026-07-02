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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { MediaTile } from "#/components/MediaTile/MediaTile";
import { externalHref } from "#/lib/url/hostname";

import type { TrustCenterReferenceListItem_reference$key } from "./__generated__/TrustCenterReferenceListItem_reference.graphql";

const trustCenterReferenceListItemFragment = graphql`
  fragment TrustCenterReferenceListItem_reference on TrustCenterReference {
    name
    websiteUrl
    logo {
      downloadUrl
    }
  }
`;

interface TrustCenterReferenceListItemProps {
  referenceKey: TrustCenterReferenceListItem_reference$key;
}

// A single "Trusted by" logo tile, linking to the reference's website.
export function TrustCenterReferenceListItem({ referenceKey }: TrustCenterReferenceListItemProps) {
  const reference = useFragment(trustCenterReferenceListItemFragment, referenceKey);

  return (
    <a
      className="block"
      href={externalHref(reference.websiteUrl)}
      target="_blank"
      rel="noopener noreferrer"
    >
      <MediaTile
        variant="logo"
        backdropSrc={reference.logo.downloadUrl}
        media={<img src={reference.logo.downloadUrl} alt={reference.name} />}
        label={(
          <Text size={2} weight="medium" color="neutral" highContrast align="center">
            {reference.name}
          </Text>
        )}
      />
    </a>
  );
}
