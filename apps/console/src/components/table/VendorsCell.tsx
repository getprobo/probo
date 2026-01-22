import { faviconUrl } from "@probo/helpers";
import { Avatar, Badge, IconCrossLargeX } from "@probo/ui";

import type { VendorGraphSelectQuery } from "#/__generated__/core/VendorGraphSelectQuery.graphql";
import { GraphQLCell } from "#/components/table/GraphQLCell";
import { vendorsSelectQuery } from "#/hooks/graph/VendorGraph";

type Vendor = {
  id: string;
  name: string;
  websiteUrl: string | null | undefined;
};

type Props = {
  name: string;
  defaultValue?: Vendor[];
  organizationId: string;
};

const empty = [] as Vendor[];

export function VendorsCell(props: Props) {
  return (
    <GraphQLCell<VendorGraphSelectQuery, Vendor>
      multiple
      name={props.name}
      query={vendorsSelectQuery}
      variables={{
        organizationId: props.organizationId,
      }}
      items={data =>
        data.organization?.vendors?.edges?.map(edge => edge.node) ?? []}
      itemRenderer={({ item, onRemove }) => (
        <VendorBadge vendor={item} onRemove={onRemove} />
      )}
      defaultValue={props.defaultValue ?? empty}
    />
  );
}

function VendorBadge({
  vendor,
  onRemove,
}: {
  vendor: Vendor;
  onRemove?: (v: Vendor) => void;
}) {
  return (
    <Badge variant="neutral" className="flex items-center gap-1">
      <Avatar name={vendor.name} src={faviconUrl(vendor.websiteUrl)} size="s" />
      <span className="max-w-[100px] text-ellipsis overflow-hidden min-w-0 block">
        {vendor.name}
      </span>
      {onRemove && (
        <button
          onClick={() => onRemove(vendor)}
          className="size-4 hover:text-txt-primary cursor-pointer"
          type="button"
        >
          <IconCrossLargeX size={14} />
        </button>
      )}
    </Badge>
  );
}
