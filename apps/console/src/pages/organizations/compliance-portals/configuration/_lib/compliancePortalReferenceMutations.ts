// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { graphql } from "react-relay";

import type { compliancePortalReferenceMutationsCreateMutation } from "#/__generated__/core/compliancePortalReferenceMutationsCreateMutation.graphql";
import type { compliancePortalReferenceMutationsDeleteMutation } from "#/__generated__/core/compliancePortalReferenceMutationsDeleteMutation.graphql";
import type { compliancePortalReferenceMutationsUpdateMutation } from "#/__generated__/core/compliancePortalReferenceMutationsUpdateMutation.graphql";
import type { compliancePortalReferenceMutationsUpdateRankMutation } from "#/__generated__/core/compliancePortalReferenceMutationsUpdateRankMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

export const createCompliancePortalReferenceMutation = graphql`
  mutation compliancePortalReferenceMutationsCreateMutation(
    $input: CreateCompliancePortalReferenceInput!
    $connections: [ID!]!
  ) {
    createCompliancePortalReference(input: $input) {
      compliancePortalReferenceEdge @appendEdge(connections: $connections) {
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

export const updateCompliancePortalReferenceMutation = graphql`
  mutation compliancePortalReferenceMutationsUpdateMutation(
    $input: UpdateCompliancePortalReferenceInput!
  ) {
    updateCompliancePortalReference(input: $input) {
      compliancePortalReference {
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

export const deleteCompliancePortalReferenceMutation = graphql`
  mutation compliancePortalReferenceMutationsDeleteMutation(
    $input: DeleteCompliancePortalReferenceInput!
    $connections: [ID!]!
  ) {
    deleteCompliancePortalReference(input: $input) {
      deletedCompliancePortalReferenceId @deleteEdge(connections: $connections)
    }
  }
`;

export function useCreateCompliancePortalReferenceMutation() {
  return useMutation<compliancePortalReferenceMutationsCreateMutation>(
    createCompliancePortalReferenceMutation,
    {
      successMessage: "Reference created successfully",
      errorToast: "Failed to create reference",
    },
  );
}

export function useUpdateCompliancePortalReferenceMutation() {
  return useMutation<compliancePortalReferenceMutationsUpdateMutation>(
    updateCompliancePortalReferenceMutation,
    {
      successMessage: "Reference updated successfully",
      errorToast: "Failed to update reference",
    },
  );
}

export const updateCompliancePortalReferenceRankMutation = graphql`
  mutation compliancePortalReferenceMutationsUpdateRankMutation(
    $input: UpdateCompliancePortalReferenceInput!
  ) {
    updateCompliancePortalReference(input: $input) {
      compliancePortalReference {
        id
        rank
      }
    }
  }
`;

export function useUpdateCompliancePortalReferenceRankMutation() {
  return useMutation<compliancePortalReferenceMutationsUpdateRankMutation>(
    updateCompliancePortalReferenceRankMutation,
    {
      successMessage: "Order updated successfully",
      errorToast: "Failed to update order",
    },
  );
}

export function useDeleteCompliancePortalReferenceMutation() {
  return useMutation<compliancePortalReferenceMutationsDeleteMutation>(
    deleteCompliancePortalReferenceMutation,
    {
      successMessage: "Reference deleted successfully",
      errorToast: "Failed to delete reference",
    },
  );
}
