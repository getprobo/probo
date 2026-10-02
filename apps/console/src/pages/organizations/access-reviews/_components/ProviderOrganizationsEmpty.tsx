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

import { IconArrowLink, IconWarning } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ProviderOrganizationsEmpty_source$key } from "#/__generated__/core/ProviderOrganizationsEmpty_source.graphql";

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

  return (
    <div className="flex max-w-80 flex-col gap-2">
      <div className="flex items-start gap-2 text-xs text-txt-tertiary">
        <IconWarning size={14} className="mt-0.5 shrink-0 text-txt-warning" />
        <div className="space-y-1">
          <p className="font-medium text-txt-primary">
            {t("accessReviewSourceRow.organizations.empty.title", {
              provider: providerName,
            })}
          </p>
          <p>{t("accessReviewSourceRow.organizations.empty.description")}</p>
          {remediationUrl && (
            <a
              href={remediationUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 underline hover:no-underline"
            >
              {t("accessReviewSourceRow.organizations.empty.remediation")}
              <IconArrowLink size={12} />
            </a>
          )}
        </div>
      </div>
      <div className="space-y-1">
        <p className="text-xs text-txt-tertiary">
          {t("accessReviewSourceRow.organizations.empty.manualLabel")}
        </p>
        <ManualOrgInput sourceKey={source} onSubmit={onSubmit} />
      </div>
    </div>
  );
}
