import { graphql } from "relay-runtime";
import {
  Card,
  Button,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  IconChevronDown,
  IconCheckmark1,
  IconCrossLargeX,
  Badge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useFragment } from "react-relay";
import { useMemo, useState } from "react";
import { sprintf } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import clsx from "clsx";
import type { TrustCenterVendorsCardFragment$key } from "./__generated__/TrustCenterVendorsCardFragment.graphql";

const trustCenterVendorFragment = graphql`
  fragment TrustCenterVendorsCardFragment on Vendor {
    id
    name
    category
    description
    showOnTrustCenter
    createdAt
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      id: string;
      showOnTrustCenter: boolean;
    } & Params;
  };
}) => void;

type Props<Params> = {
  vendors: TrustCenterVendorsCardFragment$key[];
  params: Params;
  disabled?: boolean;
  onToggleVisibility: Mutation<Params>;
  variant?: "card" | "table";
};

export function TrustCenterVendorsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(4);
  const vendors = useMemo(() => {
    return limit ? props.vendors.slice(0, limit) : props.vendors;
  }, [props.vendors, limit]);
  const showMoreButton = limit !== null && props.vendors.length > limit;
  const variant = props.variant ?? "table";

  const onToggleVisibility = (vendorId: string, showOnTrustCenter: boolean) => {
    props.onToggleVisibility({
      variables: {
        input: {
          id: vendorId,
          showOnTrustCenter,
          ...props.params,
        },
      },
    });
  };

  const Wrapper = variant === "card" ? Card : "div";

  return (
    <Wrapper padded className="space-y-[10px]">
      <Table className={clsx(variant === "card" && "bg-invert")}>
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
    </Wrapper>
  );
}

function VendorRow(props: {
  vendor: TrustCenterVendorsCardFragment$key;
  onToggleVisibility: (vendorId: string, showOnTrustCenter: boolean) => void;
  disabled?: boolean;
}) {
  const vendor = useFragment(trustCenterVendorFragment, props.vendor);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/vendors/${vendor.id}/overview`}>
      <Td>
        <div className="flex gap-4 items-center">
          {vendor.name}
        </div>
      </Td>
      <Td>
        <Badge variant="neutral">
          {vendor.category}
        </Badge>
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          {vendor.showOnTrustCenter ? (
            <>
              <IconCheckmark1 className="w-4 h-4 text-txt-primary" />
              <span className="text-txt-primary">{__("Visible")}</span>
            </>
          ) : (
            <>
              <IconCrossLargeX className="w-4 h-4 text-txt-tertiary" />
              <span className="text-txt-tertiary">{__("Hidden")}</span>
            </>
          )}
        </div>
      </Td>
      <Td noLink width={100} className="text-end">
        <Button
          variant="secondary"
          onClick={() => props.onToggleVisibility(vendor.id, !vendor.showOnTrustCenter)}
          icon={vendor.showOnTrustCenter ? IconCrossLargeX : IconCheckmark1}
          disabled={props.disabled}
        >
          {vendor.showOnTrustCenter ? __("Hide") : __("Show")}
        </Button>
      </Td>
    </Tr>
  );
}
