import { graphql } from "react-relay";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import type { TrustCenterAccessGraphQuery } from "/__generated__/core/TrustCenterAccessGraphQuery.graphql";
import type {
  TrustCenterAccessGraph_accesses$data,
  TrustCenterAccessGraph_accesses$key,
} from "/__generated__/core/TrustCenterAccessGraph_accesses.graphql";

export const trustCenterAccessesPaginationFragment = graphql`
  fragment TrustCenterAccessGraph_accesses on TrustCenter
  @refetchable(queryName: "TrustCenterAccessGraphPaginationQuery") {
    accesses(
      first: $count
      after: $cursor
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "TrustCenterAccessGraph_accesses") {
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
          pendingRequestCount
          activeCount
          canUpdate: permission(action: "core:trust-center-access:update")
          canDelete: permission(action: "core:trust-center-access:delete")
        }
      }
    }
  }
`;

export const trustCenterAccessesQuery = graphql`
  query TrustCenterAccessGraphQuery(
    $trustCenterId: ID!
    $count: Int!
    $cursor: CursorKey
  ) {
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
          pendingRequestCount
          activeCount
          canUpdate: permission(action: "core:trust-center-access:update")
          canDelete: permission(action: "core:trust-center-access:delete")
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
        pendingRequestCount
        activeCount
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

export const loadTrustCenterAccessDocumentAccessesQuery = graphql`
  query TrustCenterAccessGraphLoadDocumentAccessesQuery($accessId: ID!) {
    node(id: $accessId) {
      ... on TrustCenterAccess {
        id
        availableDocumentAccesses(
          first: 100
          orderBy: { field: CREATED_AT, direction: DESC }
        ) {
          edges {
            node {
              id
              status
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
              trustCenterFile {
                id
                name
                category
              }
            }
          }
        }
      }
    }
  }
`;

interface PaginatedData {
  data: TrustCenterAccessGraph_accesses$data | null;
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
      cursor: null,
    },
    { fetchPolicy: "store-and-network" },
  );

  const trustCenter = data?.node;

  const {
    data: paginationData,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    TrustCenterAccessGraphQuery,
    TrustCenterAccessGraph_accesses$key
  >(trustCenterAccessesPaginationFragment, trustCenter);

  if (!trustCenterId) {
    return {
      data: null,
      hasNext: false,
      loadMore: () => {},
      isLoadingNext: false,
    };
  }

  const loadMore = () => {
    loadNext(10);
  };

  return {
    data: paginationData,
    hasNext,
    loadMore,
    isLoadingNext,
  };
}
