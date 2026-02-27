import { graphql } from "react-relay";

import type { ComplianceBadgeGraphCreateMutation } from "#/__generated__/core/ComplianceBadgeGraphCreateMutation.graphql";
import type { ComplianceBadgeGraphDeleteMutation } from "#/__generated__/core/ComplianceBadgeGraphDeleteMutation.graphql";
import type { ComplianceBadgeGraphUpdateMutation } from "#/__generated__/core/ComplianceBadgeGraphUpdateMutation.graphql";
import type { ComplianceBadgeGraphUpdateRankMutation } from "#/__generated__/core/ComplianceBadgeGraphUpdateRankMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

export const createComplianceBadgeMutation = graphql`
  mutation ComplianceBadgeGraphCreateMutation(
    $input: CreateComplianceBadgeInput!
    $connections: [ID!]!
  ) {
    createComplianceBadge(input: $input) {
      complianceBadgeEdge @appendEdge(connections: $connections) {
        cursor
        node {
          id
          name
          iconUrl
          rank
          createdAt
          updatedAt
          canUpdate: permission(action: "core:compliance-badge:update")
          canDelete: permission(action: "core:compliance-badge:delete")
        }
      }
    }
  }
`;

export const updateComplianceBadgeMutation = graphql`
  mutation ComplianceBadgeGraphUpdateMutation(
    $input: UpdateComplianceBadgeInput!
  ) {
    updateComplianceBadge(input: $input) {
      complianceBadge {
        id
        name
        iconUrl
        rank
        createdAt
        updatedAt
        canUpdate: permission(action: "core:compliance-badge:update")
        canDelete: permission(action: "core:compliance-badge:delete")
      }
    }
  }
`;

export const deleteComplianceBadgeMutation = graphql`
  mutation ComplianceBadgeGraphDeleteMutation(
    $input: DeleteComplianceBadgeInput!
    $connections: [ID!]!
  ) {
    deleteComplianceBadge(input: $input) {
      deletedComplianceBadgeId @deleteEdge(connections: $connections)
    }
  }
`;

export const updateComplianceBadgeRankMutation = graphql`
  mutation ComplianceBadgeGraphUpdateRankMutation(
    $input: UpdateComplianceBadgeInput!
  ) {
    updateComplianceBadge(input: $input) {
      complianceBadge {
        id
        rank
      }
    }
  }
`;

export function useCreateComplianceBadgeMutation() {
  return useMutationWithToasts<ComplianceBadgeGraphCreateMutation>(
    createComplianceBadgeMutation,
    {
      successMessage: "Badge created successfully",
      errorMessage: "Failed to create badge",
    },
  );
}

export function useUpdateComplianceBadgeMutation() {
  return useMutationWithToasts<ComplianceBadgeGraphUpdateMutation>(
    updateComplianceBadgeMutation,
    {
      successMessage: "Badge updated successfully",
      errorMessage: "Failed to update badge",
    },
  );
}

export function useUpdateComplianceBadgeRankMutation() {
  return useMutationWithToasts<ComplianceBadgeGraphUpdateRankMutation>(
    updateComplianceBadgeRankMutation,
    {
      successMessage: "Order updated successfully",
      errorMessage: "Failed to update order",
    },
  );
}

export function useDeleteComplianceBadgeMutation() {
  return useMutationWithToasts<ComplianceBadgeGraphDeleteMutation>(
    deleteComplianceBadgeMutation,
    {
      successMessage: "Badge deleted successfully",
      errorMessage: "Failed to delete badge",
    },
  );
}
