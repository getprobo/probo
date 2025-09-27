import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { trustVendorsQuery } from "/queries/TrustGraph";
import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Table } from "@probo/ui";
import type { TrustGraphVendorsQuery } from "/queries/__generated__/TrustGraphVendorsQuery.graphql";
import { VendorRow } from "/components/VendorRow";

type Props = {
  queryRef: PreloadedQuery<TrustGraphVendorsQuery>;
};

export function SubprocessorsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(trustVendorsQuery, queryRef);
  console.log(data);
  const vendors =
    data.trustCenterBySlug?.vendors.edges.map((edge) => edge.node) ?? [];
  return (
    <div>
      <h2 className="font-medium mb-1">{__("Subprocessors")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {sprintf(
          __("Third-party subprocessors %s work with:"),
          data.trustCenterBySlug?.organization.name ?? "",
        )}
      </p>
      <Table>
        {vendors.map((vendor) => (
          <VendorRow key={vendor.id} vendor={vendor} />
        ))}
      </Table>
    </div>
  );
}
