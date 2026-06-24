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
          logo {
            downloadUrl
          }
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
        logo {
          downloadUrl
        }
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
