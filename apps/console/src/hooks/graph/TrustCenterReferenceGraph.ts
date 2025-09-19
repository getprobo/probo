import { graphql } from 'react-relay';
import { useLazyLoadQuery } from 'react-relay';
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type {
  TrustCenterReferenceGraphQuery,
  TrustCenterReferenceGraphQuery$data
} from "./__generated__/TrustCenterReferenceGraphQuery.graphql";
import type { TrustCenterReferenceGraphCreateMutation } from "./__generated__/TrustCenterReferenceGraphCreateMutation.graphql";
import type { TrustCenterReferenceGraphUpdateMutation } from "./__generated__/TrustCenterReferenceGraphUpdateMutation.graphql";
import type { TrustCenterReferenceGraphDeleteMutation } from "./__generated__/TrustCenterReferenceGraphDeleteMutation.graphql";

export const trustCenterReferencesQuery = graphql`
  query TrustCenterReferenceGraphQuery($trustCenterId: ID!) {
    node(id: $trustCenterId) {
      ... on TrustCenter {
        id
        references(first: 100, orderBy: { field: CREATED_AT, direction: DESC })
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
      trustCenterReferenceEdge @prependEdge(connections: $connections) {
        cursor
        node {
          id
          name
          description
          websiteUrl
          logoUrl
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

export function useTrustCenterReferences(trustCenterId: string): TrustCenterReferenceGraphQuery$data | null {
  const data = useLazyLoadQuery<TrustCenterReferenceGraphQuery>(
    trustCenterReferencesQuery,
    { trustCenterId: trustCenterId || "" },
    { fetchPolicy: 'network-only' }
  );

  return trustCenterId ? data : null;
}

export function useCreateTrustCenterReferenceMutation() {
  return useMutationWithToasts<TrustCenterReferenceGraphCreateMutation>(
    createTrustCenterReferenceMutation,
    {
      successMessage: "Reference created successfully",
      errorMessage: "Failed to create reference",
    }
  );
}

export function useUpdateTrustCenterReferenceMutation() {
  return useMutationWithToasts<TrustCenterReferenceGraphUpdateMutation>(
    updateTrustCenterReferenceMutation,
    {
      successMessage: "Reference updated successfully",
      errorMessage: "Failed to update reference",
    }
  );
}

export function useDeleteTrustCenterReferenceMutation() {
  return useMutationWithToasts<TrustCenterReferenceGraphDeleteMutation>(
    deleteTrustCenterReferenceMutation,
    {
      successMessage: "Reference deleted successfully",
      errorMessage: "Failed to delete reference",
    }
  );
}
