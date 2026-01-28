import { graphql } from "react-relay";

import type { TrustCenterReferenceGraphCreateMutation } from "#/__generated__/core/TrustCenterReferenceGraphCreateMutation.graphql";
import type { TrustCenterReferenceGraphDeleteMutation } from "#/__generated__/core/TrustCenterReferenceGraphDeleteMutation.graphql";
import type { TrustCenterReferenceGraphUpdateMutation } from "#/__generated__/core/TrustCenterReferenceGraphUpdateMutation.graphql";
import type { TrustCenterReferenceGraphUpdateRankMutation } from "#/__generated__/core/TrustCenterReferenceGraphUpdateRankMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

export const createTrustCenterReferenceMutation = graphql`
  mutation TrustCenterReferenceGraphCreateMutation(
    $input: CreateTrustCenterReferenceInput!
    $connections: [ID!]!
  ) {
    createTrustCenterReference(input: $input) {
      trustCenterReferenceEdge @appendEdge(connections: $connections) {
        cursor
        node {
          id
          name
          description
          websiteUrl
          logoUrl
          rank
          createdAt
          updatedAt
          canUpdate: permission(action: "core:trust-center-reference:update")
          canDelete: permission(action: "core:trust-center-reference:delete")
        }
      }
    }
  }
`;

export const updateTrustCenterReferenceMutation = graphql`
  mutation TrustCenterReferenceGraphUpdateMutation(
    $input: UpdateTrustCenterReferenceInput!
  ) {
    updateTrustCenterReference(input: $input) {
      trustCenterReference {
        id
        name
        description
        websiteUrl
        logoUrl
        rank
        createdAt
        updatedAt
        canUpdate: permission(action: "core:trust-center-reference:update")
        canDelete: permission(action: "core:trust-center-reference:delete")
      }
    }
  }
`;

export const deleteTrustCenterReferenceMutation = graphql`
  mutation TrustCenterReferenceGraphDeleteMutation(
    $input: DeleteTrustCenterReferenceInput!
    $connections: [ID!]!
  ) {
    deleteTrustCenterReference(input: $input) {
      deletedTrustCenterReferenceId @deleteEdge(connections: $connections)
    }
  }
`;

export function useCreateTrustCenterReferenceMutation() {
  return useMutationWithToasts<TrustCenterReferenceGraphCreateMutation>(
    createTrustCenterReferenceMutation,
    {
      successMessage: "Reference created successfully",
      errorMessage: "Failed to create reference",
    },
  );
}

export function useUpdateTrustCenterReferenceMutation() {
  return useMutationWithToasts<TrustCenterReferenceGraphUpdateMutation>(
    updateTrustCenterReferenceMutation,
    {
      successMessage: "Reference updated successfully",
      errorMessage: "Failed to update reference",
    },
  );
}

export const updateTrustCenterReferenceRankMutation = graphql`
  mutation TrustCenterReferenceGraphUpdateRankMutation(
    $input: UpdateTrustCenterReferenceInput!
  ) {
    updateTrustCenterReference(input: $input) {
      trustCenterReference {
        id
        rank
      }
    }
  }
`;

export function useUpdateTrustCenterReferenceRankMutation() {
  return useMutationWithToasts<TrustCenterReferenceGraphUpdateRankMutation>(
    updateTrustCenterReferenceRankMutation,
    {
      successMessage: "Order updated successfully",
      errorMessage: "Failed to update order",
    },
  );
}

export function useDeleteTrustCenterReferenceMutation() {
  return useMutationWithToasts<TrustCenterReferenceGraphDeleteMutation>(
    deleteTrustCenterReferenceMutation,
    {
      successMessage: "Reference deleted successfully",
      errorMessage: "Failed to delete reference",
    },
  );
}
