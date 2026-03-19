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
