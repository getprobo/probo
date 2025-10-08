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
      ...OverviewPageFragment
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

// Queries for custom domain (subdomain) approach
export const currentTrustGraphQuery = graphql`
  query TrustGraphCurrentQuery {
    currentTrustCenter {
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
      ...OverviewPageFragment
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

export const currentTrustDocumentsQuery = graphql`
  query TrustGraphCurrentDocumentsQuery {
    currentTrustCenter {
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

export const currentTrustVendorsQuery = graphql`
  query TrustGraphCurrentVendorsQuery {
    currentTrustCenter {
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
