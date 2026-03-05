import { graphql } from "relay-runtime";

// Queries for custom domain (subdomain) approach
export const currentTrustGraphQuery = graphql`
  query TrustGraphCurrentQuery {
    viewer {
      id
    }
    currentTrustCenter @required(action: THROW) {
      id
      slug
      viewerSubscription {
        id
        email
        createdAt
        updatedAt
      }
      logoFileUrl
      darkLogoFileUrl
      nonDisclosureAgreement {
        fileName
        fileUrl
        viewerSignature {
          status
        }
      }
      organization {
        name
        description
        websiteUrl
        email
        headquarterAddress
      }
      ...OverviewPageFragment
      vendorInfo: vendors(first: 0) {
        totalCount
      }
      audits(first: 50) {
        edges {
          node {
            id
            ...AuditRowFragment
          }
        }
      }
      complianceFrameworks(first: 50) {
        edges {
          node {
            id
            framework {
              ...FrameworkBadgeFragment
            }
          }
        }
      }
    }
  }
`;

export const currentTrustNewsQuery = graphql`
  query TrustGraphCurrentNewsQuery {
    currentTrustCenter {
      id
      complianceNews(first: 50) {
        edges {
          node {
            id
            title
            body
            updatedAt
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
      trustCenterFiles(first: 50) {
        edges {
          node {
            id
            category
            ...TrustCenterFileRowFragment
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
            countries
            ...VendorRowFragment
          }
        }
      }
    }
  }
`;
