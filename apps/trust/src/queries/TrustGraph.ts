import { graphql } from "relay-runtime";

export const trustGraphQuery = graphql`
  query TrustGraphQuery($slug: String!) {
    trustCenterBySlug(slug: $slug) {
      id
      slug
      isUserAuthenticated
      hasAcceptedNonDisclosureAgreement
      ndaFileName
      ndaFileUrl
      organization {
        name
        description
        websiteUrl
        logoUrl
        email
        headquarterAddress
      }
      ...OverviewFragment
      audits(first: 50) {
        edges {
          node {
            id
            ...AuditRowFragment
          }
        }
      }
    }
  }
`;

export const trustDocumentsQuery = graphql`
  query TrustGraphDocumentsQuery($slug: String!) {
    trustCenterBySlug(slug: $slug) {
      id
      organization {
        name
      }
      documents(first: 50) {
        edges {
          node {
            id
            documentType
            ...DocumentRowFragment
          }
        }
      }
    }
  }
`;

export const trustVendorsQuery = graphql`
  query TrustGraphVendorsQuery($slug: String!) {
    trustCenterBySlug(slug: $slug) {
      id
      organization {
        name
      }
      vendors(first: 50) {
        edges {
          node {
            id
            ...VendorRowFragment
          }
        }
      }
    }
  }
`;
