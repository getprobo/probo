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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { MediaTile } from "#/components/MediaTile/MediaTile";
import { externalHref } from "#/lib/url/hostname";

import type { TrustCenterReferenceListItem_reference$key } from "./__generated__/TrustCenterReferenceListItem_reference.graphql";

const trustCenterReferenceListItemFragment = graphql`
  fragment TrustCenterReferenceListItem_reference on TrustCenterReference @throwOnFieldError {
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
