import { graphql } from 'react-relay';
import { useLazyLoadQuery } from 'react-relay';
import type {
  TrustCenterAccessGraphQuery,
  TrustCenterAccessGraphQuery$data
} from "./__generated__/TrustCenterAccessGraphQuery.graphql";

export const trustCenterAccessesQuery = graphql`
  query TrustCenterAccessGraphQuery($trustCenterId: ID!) {
    node(id: $trustCenterId) {
      ... on TrustCenter {
        id
        accesses(first: 100, orderBy: { field: CREATED_AT, direction: DESC })
          @connection(key: "TrustCenterAccessTab_accesses") {
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
              email
              name
              active
              createdAt
            }
          }
        }
      }
    }
  }
`;

export const createTrustCenterAccessMutation = graphql`
  mutation TrustCenterAccessGraphCreateMutation(
    $input: CreateTrustCenterAccessInput!
    $connections: [ID!]!
  ) {
    createTrustCenterAccess(input: $input) {
      trustCenterAccessEdge @prependEdge(connections: $connections) {
        cursor
        node {
          id
          email
          name
          active
          createdAt
        }
      }
    }
  }
`;

export const updateTrustCenterAccessMutation = graphql`
  mutation TrustCenterAccessGraphUpdateMutation(
    $input: UpdateTrustCenterAccessInput!
  ) {
    updateTrustCenterAccess(input: $input) {
      trustCenterAccess {
        id
        email
        name
        active
        createdAt
        updatedAt
      }
    }
  }
`;

export const deleteTrustCenterAccessMutation = graphql`
  mutation TrustCenterAccessGraphDeleteMutation(
    $input: DeleteTrustCenterAccessInput!
    $connections: [ID!]!
  ) {
    deleteTrustCenterAccess(input: $input) {
      deletedTrustCenterAccessId @deleteEdge(connections: $connections)
    }
  }
`;

export function useTrustCenterAccesses(trustCenterId: string): TrustCenterAccessGraphQuery$data | null {
  // Always call useLazyLoadQuery to maintain consistent hook order
  // Use a placeholder value when trustCenterId is empty
  const data = useLazyLoadQuery<TrustCenterAccessGraphQuery>(
    trustCenterAccessesQuery,
    { trustCenterId: trustCenterId || "" },
    { fetchPolicy: 'store-and-network' }
  );

  // Return null if trustCenterId was empty, otherwise return the data
  return trustCenterId ? data : null;
}
