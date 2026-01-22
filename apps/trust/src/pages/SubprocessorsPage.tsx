import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";

import { currentTrustVendorsQuery } from "/queries/TrustGraph";
import type { TrustGraphCurrentVendorsQuery } from "/queries/__generated__/TrustGraphCurrentVendorsQuery.graphql";
import { VendorRow } from "/components/VendorRow";
import { Rows } from "/components/Rows.tsx";

type Props = {
  queryRef: PreloadedQuery<TrustGraphCurrentVendorsQuery>;
};

export function SubprocessorsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(currentTrustVendorsQuery, queryRef);
  const vendors
    = data.currentTrustCenter?.vendors.edges.map(edge => edge.node) ?? [];

  const hasAnyCountries = vendors.some(vendor => vendor.countries.length > 0);

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
        {vendors.map(vendor => (
          <VendorRow key={vendor.id} vendor={vendor} hasAnyCountries={hasAnyCountries} />
        ))}
      </Rows>
    </div>
  );
}
