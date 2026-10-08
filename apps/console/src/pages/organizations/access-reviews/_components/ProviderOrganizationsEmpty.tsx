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

import { ArrowSquareOutIcon, WarningCircleIcon } from "@phosphor-icons/react";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ProviderOrganizationsEmpty_source$key } from "#/__generated__/core/ProviderOrganizationsEmpty_source.graphql";

import { organizationsEmpty } from "../sources/_components/variants";

import { ManualOrgInput } from "./ManualOrgInput";

const organizationsEmptyFragment = graphql`
  fragment ProviderOrganizationsEmpty_source on AccessReviewSource {
    providerOrganizations {
      remediationUrl
    }
    ...ManualOrgInput_source
  }
`;

interface ProviderOrganizationsEmptyProps {
  sourceKey: ProviderOrganizationsEmpty_source$key;
  providerName: string;
  onSubmit: (slug: string) => void;
}

// The provider answered, with nothing. Typing a slug does not fix the usual
// causes (unapproved app, personal account), so the explanation leads and the
// manual input stays available underneath it.
export function ProviderOrganizationsEmpty({
  sourceKey,
  providerName,
  onSubmit,
}: ProviderOrganizationsEmptyProps) {
  const { t } = useTranslation();
  const source = useFragment(organizationsEmptyFragment, sourceKey);
  const remediationUrl = source.providerOrganizations.remediationUrl;
  const { root, copy } = organizationsEmpty();

  return (
    <div className={root()}>
      <Callout
        size={1}
        variant="soft"
        color="amber"
        icon={<WarningCircleIcon weight="fill" />}
      >
        <div className={copy()}>
          <Heading level={3} size={2} weight="medium" color="current">
            {t("accessReviewSourceRow.organizations.empty.title", {
              provider: providerName,
            })}
          </Heading>
          <Text size={1} color="current">
            {t("accessReviewSourceRow.organizations.empty.description")}
          </Text>
          {remediationUrl && (
            <Anchor
              href={remediationUrl}
              target="_blank"
              rel="noopener noreferrer"
              size={1}
              color="amber"
              highContrast
              iconEnd={<ArrowSquareOutIcon />}
            >
              {t("accessReviewSourceRow.organizations.empty.remediation")}
            </Anchor>
          )}
        </div>
      </Callout>
      <ManualOrgInput sourceKey={source} onSubmit={onSubmit} />
    </div>
  );
}
