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
              hasAcceptedNonDisclosureAgreement
              createdAt
              documentAccesses(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
                edges {
                  node {
                    id
                    active
                    createdAt
                    updatedAt
                    document {
                      id
                      title
                      documentType
                    }
                    report {
                      id
                      filename
                      audit {
                        id
                        framework {
                          name
                        }
                      }
                    }
                  }
                }
              }
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
          hasAcceptedNonDisclosureAgreement
          createdAt
          documentAccesses(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
            edges {
              node {
                id
                active
                createdAt
                updatedAt
                document {
                  id
                  title
                  documentType
                }
            report {
              id
              filename
              audit {
                id
                framework {
                  name
                }
              }
            }
              }
            }
          }
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
        hasAcceptedNonDisclosureAgreement
        createdAt
        updatedAt
        documentAccesses(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              active
              createdAt
              updatedAt
              document {
                id
                title
                documentType
              }
            report {
              id
              filename
              audit {
                id
                framework {
                  name
                }
              }
            }
            }
          }
        }
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

export const updateTrustCenterDocumentAccessStatusMutation = graphql`
  mutation TrustCenterAccessGraphUpdateDocumentAccessStatusMutation(
    $input: UpdateTrustCenterDocumentAccessStatusInput!
  ) {
    updateTrustCenterDocumentAccessStatus(input: $input) {
      trustCenterDocumentAccess {
        id
        active
        updatedAt
        document {
          id
          title
          documentType
        }
            report {
              id
              filename
              audit {
                id
                framework {
                  name
                }
              }
            }
      }
    }
  }
`;

export function useTrustCenterAccesses(trustCenterId: string): TrustCenterAccessGraphQuery$data | null {
  // Always call useLazyLoadQuery to maintain consistent hook order
  // Use a placeholder value when trustCenterId is empty
  const data = useLazyLoadQuery<TrustCenterAccessGraphQuery>(
    trustCenterAccessesQuery,
    { trustCenterId: trustCenterId || "" },
    { fetchPolicy: 'network-only' }
  );

  // Return null if trustCenterId was empty, otherwise return the data
  return trustCenterId ? data : null;
}
