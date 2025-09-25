import { graphql } from 'react-relay';
import { useLazyLoadQuery, usePaginationFragment } from 'react-relay';
import type {
  TrustCenterAccessGraphQuery
} from "./__generated__/TrustCenterAccessGraphQuery.graphql";

export const trustCenterAccessesPaginationFragment = graphql`
  fragment TrustCenterAccessGraph_accesses on TrustCenter
  @refetchable(queryName: "TrustCenterAccessGraphPaginationQuery") {
    accesses(first: $count, after: $cursor, orderBy: { field: CREATED_AT, direction: DESC })
      @connection(key: "TrustCenterAccessGraph_accesses") {
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
`;

export const trustCenterAccessesQuery = graphql`
  query TrustCenterAccessGraphQuery($trustCenterId: ID!, $count: Int!, $cursor: CursorKey) {
    node(id: $trustCenterId) {
      ... on TrustCenter {
        id
        ...TrustCenterAccessGraph_accesses
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

interface PaginatedData {
  data: { node: any } | null;
  hasNext: boolean;
  loadMore: () => void;
  isLoadingNext: boolean;
}

export function useTrustCenterAccesses(trustCenterId: string): PaginatedData {
  const data = useLazyLoadQuery<TrustCenterAccessGraphQuery>(
    trustCenterAccessesQuery,
    {
      trustCenterId: trustCenterId || "",
      count: 10,
      cursor: null
    },
    { fetchPolicy: 'network-only' }
  );

  if (!trustCenterId) {
    return {
      data: null,
      hasNext: false,
      loadMore: () => {},
      isLoadingNext: false,
    };
  }

  const trustCenter = data?.node as any;

  const {
    data: paginationData,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment(
    trustCenterAccessesPaginationFragment,
    trustCenter
  );

  const loadMore = () => {
    loadNext(10);
  };

  return {
    data: { node: paginationData },
    hasNext,
    loadMore,
    isLoadingNext,
  };
}
