import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconCheckmark1,
  IconChevronDown,
  IconCrossLargeX,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useMemo, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { TrustCenterVendorsCardFragment$key } from "/__generated__/core/TrustCenterVendorsCardFragment.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";

const trustCenterVendorFragment = graphql`
  fragment TrustCenterVendorsCardFragment on Vendor {
    id
    name
    category
    description
    showOnTrustCenter
    createdAt
    canUpdate: permission(action: "core:vendor:update")
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      id: string;
      showOnTrustCenter: boolean;
    } & Params;
  };
}) => Promise<void>;

type Props<Params> = {
  vendors: TrustCenterVendorsCardFragment$key[];
  params: Params;
  disabled?: boolean;
  onToggleVisibility: Mutation<Params>;
};

export function TrustCenterVendorsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(100);
  const vendors = useMemo(() => {
    return limit ? props.vendors.slice(0, limit) : props.vendors;
  }, [props.vendors, limit]);
  const showMoreButton = limit !== null && props.vendors.length > limit;

  const onToggleVisibility = async (vendorId: string, showOnTrustCenter: boolean) => {
    await props.onToggleVisibility({
      variables: {
        input: {
          id: vendorId,
          showOnTrustCenter,
          ...props.params,
        },
      },
    });
  };

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Category")}</Th>
            <Th>{__("Visibility")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {vendors.length === 0 && (
            <Tr>
              <Td colSpan={4} className="text-center text-txt-secondary">
                {__("No vendors available")}
              </Td>
            </Tr>
          )}
          {vendors.map((vendor, index) => (
            <VendorRow
              key={index}
              vendor={vendor}
              onToggleVisibility={onToggleVisibility}
              disabled={props.disabled}
            />
          ))}
        </Tbody>
      </Table>
      {showMoreButton && (
        <Button
          variant="tertiary"
          onClick={() => setLimit(null)}
          className="mt-3 mx-auto"
          icon={IconChevronDown}
        >
          {sprintf(__("Show %s more"), props.vendors.length - limit)}
        </Button>
      )}
    </div>
  );
}

function VendorRow(props: {
  vendor: TrustCenterVendorsCardFragment$key;
  onToggleVisibility: (vendorId: string, showOnTrustCenter: boolean) => Promise<void>;
  disabled?: boolean;
}) {
  const vendor = useFragment(trustCenterVendorFragment, props.vendor);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/vendors/${vendor.id}/overview`}>
      <Td>
        <div className="flex gap-4 items-center">{vendor.name}</div>
      </Td>
      <Td>
        <Badge variant="neutral">{vendor.category}</Badge>
      </Td>
      <Td>
        <Badge variant={vendor.showOnTrustCenter ? "success" : "danger"}>
          {vendor.showOnTrustCenter ? __("Visible") : __("None")}
        </Badge>
      </Td>
      <Td noLink width={100} className="text-end">
        {vendor.canUpdate && (
          <Button
            variant="secondary"
            onClick={() =>
              void props.onToggleVisibility(vendor.id, !vendor.showOnTrustCenter)}
            icon={vendor.showOnTrustCenter ? IconCrossLargeX : IconCheckmark1}
            disabled={props.disabled}
          >
            {vendor.showOnTrustCenter ? __("Hide") : __("Show")}
          </Button>
        )}
      </Td>
    </Tr>
  );
}
