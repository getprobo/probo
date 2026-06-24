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
