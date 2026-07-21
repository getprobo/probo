// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";

import { Rows } from "#/components/Rows";
import { SubprocessorRow } from "#/components/SubprocessorRow";
import type { CompliancePortalGraphCurrentSubprocessorsQuery } from "#/queries/__generated__/CompliancePortalGraphCurrentSubprocessorsQuery.graphql";
import { currentTrustSubprocessorsQuery } from "#/queries/CompliancePortalGraph";

type Props = {
  queryRef: PreloadedQuery<CompliancePortalGraphCurrentSubprocessorsQuery>;
};

export function SubprocessorsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery<CompliancePortalGraphCurrentSubprocessorsQuery>(currentTrustSubprocessorsQuery, queryRef);
  const subprocessors
    = data.currentCompliancePortal?.subprocessors.edges.map(edge => edge.node) ?? [];

  const hasAnyCountries = subprocessors.some(subprocessor => subprocessor.countries.length > 0);

  return (
    <div>
      <h2 className="font-medium mb-1">{__("Subprocessors")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {sprintf(
          __("Third-party subprocessors %s work with:"),
          data.currentCompliancePortal?.entityName ?? "",
        )}
      </p>
      <Rows>
        {subprocessors.map(subprocessor => (
          <SubprocessorRow key={subprocessor.id} subprocessor={subprocessor} hasAnyCountries={hasAnyCountries} />
        ))}
      </Rows>
    </div>
  );
}
