// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
      externalUrls(first: 20) {
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

export const currentTrustSubprocessorsQuery = graphql`
  query TrustGraphCurrentSubprocessorsQuery {
    currentTrustCenter {
      id
      organization {
        name
      }
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
