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

import type { compliancePageReferenceMutationsCreateMutation } from "#/__generated__/core/compliancePageReferenceMutationsCreateMutation.graphql";
import type { compliancePageReferenceMutationsDeleteMutation } from "#/__generated__/core/compliancePageReferenceMutationsDeleteMutation.graphql";
import type { compliancePageReferenceMutationsUpdateMutation } from "#/__generated__/core/compliancePageReferenceMutationsUpdateMutation.graphql";
import type { compliancePageReferenceMutationsUpdateRankMutation } from "#/__generated__/core/compliancePageReferenceMutationsUpdateRankMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

export const createCompliancePageReferenceMutation = graphql`
  mutation compliancePageReferenceMutationsCreateMutation(
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

export const updateCompliancePageReferenceMutation = graphql`
  mutation compliancePageReferenceMutationsUpdateMutation(
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

export const deleteCompliancePageReferenceMutation = graphql`
  mutation compliancePageReferenceMutationsDeleteMutation(
    $input: DeleteCompliancePortalReferenceInput!
    $connections: [ID!]!
  ) {
    deleteCompliancePortalReference(input: $input) {
      deletedCompliancePortalReferenceId @deleteEdge(connections: $connections)
    }
  }
`;

export function useCreateCompliancePageReferenceMutation() {
  return useMutation<compliancePageReferenceMutationsCreateMutation>(
    createCompliancePageReferenceMutation,
    {
      successMessage: "Reference created successfully",
      errorToast: "Failed to create reference",
    },
  );
}

export function useUpdateCompliancePageReferenceMutation() {
  return useMutation<compliancePageReferenceMutationsUpdateMutation>(
    updateCompliancePageReferenceMutation,
    {
      successMessage: "Reference updated successfully",
      errorToast: "Failed to update reference",
    },
  );
}

export const updateCompliancePageReferenceRankMutation = graphql`
  mutation compliancePageReferenceMutationsUpdateRankMutation(
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

export function useUpdateCompliancePageReferenceRankMutation() {
  return useMutation<compliancePageReferenceMutationsUpdateRankMutation>(
    updateCompliancePageReferenceRankMutation,
    {
      successMessage: "Order updated successfully",
      errorToast: "Failed to update order",
    },
  );
}

export function useDeleteCompliancePageReferenceMutation() {
  return useMutation<compliancePageReferenceMutationsDeleteMutation>(
    deleteCompliancePageReferenceMutation,
    {
      successMessage: "Reference deleted successfully",
      errorToast: "Failed to delete reference",
    },
  );
}
