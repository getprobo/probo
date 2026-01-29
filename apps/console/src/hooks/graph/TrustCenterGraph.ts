import { graphql } from "relay-runtime";

import type { TrustCenterGraphUpdateMutation } from "#/__generated__/core/TrustCenterGraphUpdateMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const updateTrustCenterMutation = graphql`
  mutation TrustCenterGraphUpdateMutation($input: UpdateTrustCenterInput!) {
    updateTrustCenter(input: $input) {
      trustCenter {
        id
        active
        updatedAt
      }
    }
  }
`;

export function useUpdateTrustCenterMutation() {
  return useMutationWithToasts<TrustCenterGraphUpdateMutation>(
    updateTrustCenterMutation,
    {
      successMessage: "Compliance Page updated successfully",
      errorMessage: "Failed to update compliance page",
    },
  );
}

const uploadTrustCenterNDAMutation = graphql`
  mutation TrustCenterGraphUploadNDAMutation(
    $input: UploadTrustCenterNDAInput!
  ) {
    uploadTrustCenterNDA(input: $input) {
      trustCenter {
        id
        ndaFileName
        updatedAt
      }
    }
  }
`;

export function useUploadTrustCenterNDAMutation() {
  return useMutationWithToasts(uploadTrustCenterNDAMutation, {
    successMessage: "NDA uploaded successfully",
    errorMessage: "Failed to upload NDA",
  });
}

const deleteTrustCenterNDAMutation = graphql`
  mutation TrustCenterGraphDeleteNDAMutation(
    $input: DeleteTrustCenterNDAInput!
  ) {
    deleteTrustCenterNDA(input: $input) {
      trustCenter {
        id
        ndaFileName
        updatedAt
      }
    }
  }
`;

export function useDeleteTrustCenterNDAMutation() {
  return useMutationWithToasts(deleteTrustCenterNDAMutation, {
    successMessage: "NDA deleted successfully",
    errorMessage: "Failed to delete NDA",
  });
}
