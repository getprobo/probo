import { graphql } from "react-relay";

export const updateTrustCenterAccessMutation = graphql`
  mutation TrustCenterAccessGraphUpdateMutation(
    $input: UpdateTrustCenterAccessInput!
  ) {
    updateTrustCenterAccess(input: $input) {
      trustCenterAccess {
        id
        email
        name
        state
        hasAcceptedNonDisclosureAgreement
        createdAt
        updatedAt
        pendingRequestCount
        activeCount
      }
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
