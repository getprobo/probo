// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";

import { Rows } from "#/components/Rows";
import { SubprocessorRow } from "#/components/SubprocessorRow";
import type { TrustGraphCurrentSubprocessorsQuery } from "#/queries/__generated__/TrustGraphCurrentSubprocessorsQuery.graphql";
import { currentTrustSubprocessorsQuery } from "#/queries/TrustGraph";

type Props = {
  queryRef: PreloadedQuery<TrustGraphCurrentSubprocessorsQuery>;
};

export function SubprocessorsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(currentTrustSubprocessorsQuery, queryRef);
  const subprocessors
    = data.currentTrustCenter?.subprocessors.edges.map(edge => edge.node) ?? [];

  const hasAnyCountries = subprocessors.some(subprocessor => subprocessor.countries.length > 0);

  return (
    <div>
      <h2 className="font-medium mb-1">{__("Subprocessors")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {sprintf(
          __("Third-party subprocessors %s work with:"),
          data.currentTrustCenter?.organization.name ?? "",
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
