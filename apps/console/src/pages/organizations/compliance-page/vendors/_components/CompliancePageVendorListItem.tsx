// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconCheckmark1, IconCrossLargeX, Td, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageVendorListItem_vendorFragment$key } from "#/__generated__/core/CompliancePageVendorListItem_vendorFragment.graphql";
import type { CompliancePageVendorListItemMutation } from "#/__generated__/core/CompliancePageVendorListItemMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const vendorFragment = graphql`
  fragment CompliancePageVendorListItem_vendorFragment on Vendor {
    id
    category
    name
    showOnTrustCenter
    canUpdate: permission(action: "core:vendor:update")
  }
`;

const updateVendorVisibilityMutation = graphql`
  mutation CompliancePageVendorListItemMutation($input: UpdateVendorInput!) {
    updateVendor(input: $input) {
      vendor {
        id
        showOnTrustCenter
        ...CompliancePageVendorListItem_vendorFragment
      }
    }
  }
`;

export function CompliancePageVendorListItem(props: {
  vendorFragmentRef: CompliancePageVendorListItem_vendorFragment$key;
}) {
  const { vendorFragmentRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const vendor = useFragment<CompliancePageVendorListItem_vendorFragment$key>(
    vendorFragment,
    vendorFragmentRef,
  );
  const [updateVendorVisibility, isUpadtingVendorVisibility] = useMutationWithToasts<
    CompliancePageVendorListItemMutation
  >(
    updateVendorVisibilityMutation,
    {
      successMessage: __("Subprocessor visibility updated successfully."),
      errorMessage: __("Failed to update subprocessor visibility"),
    },
  );

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
              void updateVendorVisibility({
                variables: {
                  input: {
                    id: vendor.id,
                    showOnTrustCenter: !vendor.showOnTrustCenter,
                  },
                },
              })}
            icon={vendor.showOnTrustCenter ? IconCrossLargeX : IconCheckmark1}
            disabled={isUpadtingVendorVisibility}
          >
            {vendor.showOnTrustCenter ? __("Hide") : __("Show")}
          </Button>
        )}
      </Td>
    </Tr>
  );
};
