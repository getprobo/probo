import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts.ts";
import type { VendorGraphCreateMutation } from "./__generated__/VendorGraphCreateMutation.graphql.ts";
import type { VendorGraphDeleteMutation } from "./__generated__/VendorGraphDeleteMutation.graphql.ts";

const createVendorMutation = graphql`
  mutation VendorGraphCreateMutation(
    $input: CreateVendorInput!
    $connections: [ID!]!
  ) {
    createVendor(input: $input) {
      vendorEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          websiteUrl
          createdAt
          updatedAt
        }
      }
    }
  }
`;

export function useCreateVendorMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<VendorGraphCreateMutation>(
    createVendorMutation,
    {
      successMessage: __("Vendor created successfully."),
      errorMessage: __("Failed to create vendor. Please try again."),
    },
  );
}

const deleteVendorMutation = graphql`
  mutation VendorGraphDeleteMutation(
    $input: DeleteVendorInput!
    $connections: [ID!]!
  ) {
    deleteVendor(input: $input) {
      deletedVendorId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteVendorMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<VendorGraphDeleteMutation>(
    deleteVendorMutation,
    {
      successMessage: __("Vendor deleted successfully."),
      errorMessage: __("Failed to delete vendor. Please try again."),
    },
  );
}
