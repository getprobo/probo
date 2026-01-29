import { graphql } from "relay-runtime";

// Queries for custom domain (subdomain) approach
export const currentTrustGraphQuery = graphql`
  query TrustGraphCurrentQuery {
    viewer {
      email
      fullName
    }
    currentTrustCenter {
      id
      slug
      isViewerMember
      hasAcceptedNonDisclosureAgreement
      logoFileUrl
      darkLogoFileUrl
      ndaFileName
      ndaFileUrl
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
