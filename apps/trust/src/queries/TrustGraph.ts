// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { graphql } from "relay-runtime";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

// Queries for custom domain (subdomain) approach
export const currentTrustGraphQuery = graphql`
  query TrustGraphCurrentQuery {
    viewer {
      id
    }
    currentTrustCenter @required(action: THROW) {
      id
      slug
      title
      description
      websiteUrl
      email
      headquarterAddress
      viewerSubscription {
        id
        email
        createdAt
        updatedAt
      }
      logo {
        downloadUrl
      }
      darkLogo {
        downloadUrl
      }
      nonDisclosureAgreement {
        fileName
        fileUrl
        viewerSignature {
          status
        }
      }
      customLinks(first: 20) {
        edges {
          node {
            id
            name
            url
          }
        }
      }
      ...OverviewPageFragment
      subprocessorInfo: subprocessors(first: 0) {
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

export const currentTrustDocumentsQuery = graphql`
  query TrustGraphCurrentDocumentsQuery {
    currentTrustCenter {
      id
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

export const currentTrustSubprocessorsQuery = graphql`
  query TrustGraphCurrentSubprocessorsQuery {
    currentTrustCenter {
      id
      subprocessors(first: 50) {
        edges {
          node {
            id
            countries
            ...SubprocessorRowFragment
          }
        }
      }
    }
  }
`;
