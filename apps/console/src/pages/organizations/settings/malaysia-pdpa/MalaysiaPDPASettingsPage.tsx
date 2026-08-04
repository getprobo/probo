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

import { usePageTitle } from "@probo/hooks";
import { Badge } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { MalaysiaPDPASettingsPageQuery } from "#/__generated__/core/MalaysiaPDPASettingsPageQuery.graphql";

import { MalaysiaPDPAProfileForm } from "./_components/MalaysiaPDPAProfileForm";

export const malaysiaPDPASettingsPageQuery = graphql`
  query MalaysiaPDPASettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...MalaysiaPDPAProfileForm_organization
      }
    }
  }
`;

interface MalaysiaPDPASettingsPageProps {
  queryRef: PreloadedQuery<MalaysiaPDPASettingsPageQuery>;
}

export default function MalaysiaPDPASettingsPage({
  queryRef,
}: MalaysiaPDPASettingsPageProps) {
  const { t } = useTranslation("organizations/settings");
  const data = usePreloadedQuery<MalaysiaPDPASettingsPageQuery>(
    malaysiaPDPASettingsPageQuery,
    queryRef,
  );

  usePageTitle(t("malaysiaPDPA.page.title"));

  if (data.organization?.__typename !== "Organization") {
    throw new Error("PAGE_NOT_FOUND: organization not found");
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <h2 className="text-base font-medium">
            {t("malaysiaPDPA.page.title")}
          </h2>
          <p className="text-sm text-txt-tertiary max-w-3xl">
            {t("malaysiaPDPA.page.description")}
          </p>
        </div>
        <Badge variant="info">{t("malaysiaPDPA.page.jurisdiction")}</Badge>
      </div>

      <MalaysiaPDPAProfileForm organizationKey={data.organization} />
    </div>
  );
}
