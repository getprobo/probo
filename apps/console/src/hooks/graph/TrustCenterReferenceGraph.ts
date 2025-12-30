import { graphql } from "react-relay";
import { useLazyLoadQuery } from "react-relay";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type {
  TrustCenterReferenceGraphQuery,
  TrustCenterReferenceGraphQuery$data,
} from "/__generated__/core/TrustCenterReferenceGraphQuery.graphql";
import type { TrustCenterReferenceGraphCreateMutation } from "/__generated__/core/TrustCenterReferenceGraphCreateMutation.graphql";
import type { TrustCenterReferenceGraphUpdateMutation } from "/__generated__/core/TrustCenterReferenceGraphUpdateMutation.graphql";
import type { TrustCenterReferenceGraphUpdateRankMutation } from "/__generated__/core/TrustCenterReferenceGraphUpdateRankMutation.graphql";
import type { TrustCenterReferenceGraphDeleteMutation } from "/__generated__/core/TrustCenterReferenceGraphDeleteMutation.graphql";

export const trustCenterReferencesQuery = graphql`
  query TrustCenterReferenceGraphQuery($trustCenterId: ID!) {
    node(id: $trustCenterId) {
      ... on TrustCenter {
        id
        references(first: 100, orderBy: { field: RANK, direction: ASC })
          @connection(key: "TrustCenterReferencesSection_references") {
          __id
          pageInfo {
            hasNextPage
            hasPreviousPage
            startCursor
            endCursor
          }
          edges {
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
            }
          }
        }
      }
    }
  }
`;

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

export function useTrustCenterReferences(
  trustCenterId: string,
  refetchKey = 0,
): TrustCenterReferenceGraphQuery$data | null {
  const data = useLazyLoadQuery<TrustCenterReferenceGraphQuery>(
    trustCenterReferencesQuery,
    { trustCenterId: trustCenterId || "" },
    { fetchPolicy: "network-only", fetchKey: refetchKey },
  );

  return trustCenterId ? data : null;
}

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
