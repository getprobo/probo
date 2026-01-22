import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";

import type { TrustCenterVendorGraphUpdateMutation } from "#/__generated__/core/TrustCenterVendorGraphUpdateMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

export const trustCenterVendorUpdateMutation = graphql`
  mutation TrustCenterVendorGraphUpdateMutation($input: UpdateVendorInput!) {
    updateVendor(input: $input) {
      vendor {
        id
        showOnTrustCenter
        ...TrustCenterVendorsCardFragment
      }
    }
  }
`;

export function useTrustCenterVendorUpdate() {
  const { __ } = useTranslate();

  return useMutationWithToasts<TrustCenterVendorGraphUpdateMutation>(
    trustCenterVendorUpdateMutation,
    {
      successMessage: __("Vendor visibility updated successfully."),
      errorMessage: __("Failed to update vendor visibility"),
    },
  );
}
