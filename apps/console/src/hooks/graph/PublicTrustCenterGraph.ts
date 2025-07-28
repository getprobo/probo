import { graphql } from "relay-runtime";

export const publicTrustCenterQuery = graphql`
  query PublicTrustCenterGraphQuery($slug: String!) {
    trustCenterBySlug(slug: $slug) {
      id
      active
      slug
      organization {
        id
        name
        logoUrl
      }
      publicDocuments(first: 100) {
        edges {
          node {
            id
            title
            documentType
            versions(first: 1) {
              edges {
                node {
                  id
                  status
                }
              }
            }
          }
        }
      }
      publicAudits(first: 100) {
        edges {
          node {
            id
            framework {
              name
            }
            validFrom
            validUntil
            state
            createdAt
            report {
              id
              filename
              downloadUrl
            }
            reportUrl
          }
        }
      }
      publicVendors(first: 100) {
        edges {
          node {
            id
            name
            category
            description
            createdAt
            websiteUrl
            privacyPolicyUrl
          }
        }
      }
    }
  }
`;
