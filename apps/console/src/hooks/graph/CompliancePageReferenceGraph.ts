// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { graphql } from "react-relay";

import type { CompliancePageReferenceGraphCreateMutation } from "#/__generated__/core/CompliancePageReferenceGraphCreateMutation.graphql";
import type { CompliancePageReferenceGraphDeleteMutation } from "#/__generated__/core/CompliancePageReferenceGraphDeleteMutation.graphql";
import type { CompliancePageReferenceGraphUpdateMutation } from "#/__generated__/core/CompliancePageReferenceGraphUpdateMutation.graphql";
import type { CompliancePageReferenceGraphUpdateRankMutation } from "#/__generated__/core/CompliancePageReferenceGraphUpdateRankMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

export const createCompliancePageReferenceMutation = graphql`
  mutation CompliancePageReferenceGraphCreateMutation(
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
          logo {
            downloadUrl
          }
          rank
          createdAt
          updatedAt
          canUpdate: permission(action: "compliance-portal:portal-reference:update")
          canDelete: permission(action: "compliance-portal:portal-reference:delete")
        }
      }
    }
  }
`;

export const updateCompliancePageReferenceMutation = graphql`
  mutation CompliancePageReferenceGraphUpdateMutation(
    $input: UpdateTrustCenterReferenceInput!
  ) {
    updateTrustCenterReference(input: $input) {
      trustCenterReference {
        id
        name
        description
        websiteUrl
        logo {
          downloadUrl
        }
        rank
        createdAt
        updatedAt
        canUpdate: permission(action: "compliance-portal:portal-reference:update")
        canDelete: permission(action: "compliance-portal:portal-reference:delete")
      }
    }
  }
`;

export const deleteCompliancePageReferenceMutation = graphql`
  mutation CompliancePageReferenceGraphDeleteMutation(
    $input: DeleteTrustCenterReferenceInput!
    $connections: [ID!]!
  ) {
    deleteTrustCenterReference(input: $input) {
      deletedTrustCenterReferenceId @deleteEdge(connections: $connections)
    }
  }
`;

export function useCreateCompliancePageReferenceMutation() {
  return useMutation<CompliancePageReferenceGraphCreateMutation>(
    createCompliancePageReferenceMutation,
    {
      successMessage: "Reference created successfully",
      errorToast: "Failed to create reference",
    },
  );
}

export function useUpdateCompliancePageReferenceMutation() {
  return useMutation<CompliancePageReferenceGraphUpdateMutation>(
    updateCompliancePageReferenceMutation,
    {
      successMessage: "Reference updated successfully",
      errorToast: "Failed to update reference",
    },
  );
}

export const updateCompliancePageReferenceRankMutation = graphql`
  mutation CompliancePageReferenceGraphUpdateRankMutation(
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

export function useUpdateCompliancePageReferenceRankMutation() {
  return useMutation<CompliancePageReferenceGraphUpdateRankMutation>(
    updateCompliancePageReferenceRankMutation,
    {
      successMessage: "Order updated successfully",
      errorToast: "Failed to update order",
    },
  );
}

export function useDeleteCompliancePageReferenceMutation() {
  return useMutation<CompliancePageReferenceGraphDeleteMutation>(
    deleteCompliancePageReferenceMutation,
    {
      successMessage: "Reference deleted successfully",
      errorToast: "Failed to delete reference",
    },
  );
}
