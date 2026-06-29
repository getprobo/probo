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

import { CertificateIcon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { MediaTile } from "#/components/MediaTile/MediaTile";

import type { ComplianceFrameworkListItem_complianceFramework$key } from "./__generated__/ComplianceFrameworkListItem_complianceFramework.graphql";

const complianceFrameworkListItemFragment = graphql`
  fragment ComplianceFrameworkListItem_complianceFramework on ComplianceFramework {
    framework {
      name
      themedLogoUrl
    }
  }
`;

interface ComplianceFrameworkListItemProps {
  complianceFrameworkKey: ComplianceFrameworkListItem_complianceFramework$key;
}

// A single framework tile in the Compliance grid.
export function ComplianceFrameworkListItem({ complianceFrameworkKey }: ComplianceFrameworkListItemProps) {
  const { framework } = useFragment(complianceFrameworkListItemFragment, complianceFrameworkKey);
  const logoUrl = framework.themedLogoUrl;

  return (
    <MediaTile
      media={
        logoUrl != null && logoUrl !== ""
          ? <img src={logoUrl} alt={framework.name} />
          : <CertificateIcon size={48} weight="duotone" className="text-gold-9" />
      }
      label={(
        <Text size={2} weight="medium" color="neutral" highContrast align="center">
          {framework.name}
        </Text>
      )}
    />
  );
}
