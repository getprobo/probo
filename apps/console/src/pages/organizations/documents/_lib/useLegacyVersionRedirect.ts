// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useEffect } from "react";
import { useRelayEnvironment } from "react-relay";
import { useSearchParams } from "react-router";
import { fetchQuery, graphql } from "relay-runtime";

import type { useLegacyVersionRedirectQuery } from "#/__generated__/core/useLegacyVersionRedirectQuery.graphql";

import { documentRevisionParam, legacyDocumentVersionParam } from "./useDocumentRevisionParam";

const legacyVersionRedirectQuery = graphql`
  query useLegacyVersionRedirectQuery($documentId: ID!) {
    node(id: $documentId) {
      __typename
      ... on Document {
        history(first: 100, orderBy: { field: ENDED_AT, direction: DESC }, filter: { versioned: true }) {
          edges {
            node {
              id
              versions {
                id
              }
            }
          }
        }
      }
    }
  }
`;

// A version is a tag on a revision of the history, so a link to
// `?version=<id>` becomes `?revision=<id of the tagged revision>`. A version
// that no longer tags a revision opens the live document.
export function useLegacyVersionRedirect(documentId: string | undefined) {
  const environment = useRelayEnvironment();
  const [searchParams, setSearchParams] = useSearchParams();
  const legacyVersionId = searchParams.get(legacyDocumentVersionParam);

  useEffect(() => {
    if (documentId == null || legacyVersionId == null) {
      return;
    }

    let cancelled = false;
    const resolve = (revisionId: string | null) => {
      if (cancelled) {
        return;
      }
      setSearchParams((prev) => {
        const params = new URLSearchParams(prev);
        params.delete(legacyDocumentVersionParam);
        if (revisionId != null && !params.has(documentRevisionParam)) {
          params.set(documentRevisionParam, revisionId);
        }
        return params;
      }, { replace: true });
    };

    fetchQuery<useLegacyVersionRedirectQuery>(
      environment,
      legacyVersionRedirectQuery,
      { documentId },
      { fetchPolicy: "network-only" },
    ).toPromise().then((data) => {
      const edges = data?.node?.__typename === "Document" ? data.node.history.edges : [];
      const match = edges.find(({ node }) => node.versions.some(version => version.id === legacyVersionId));
      resolve(match?.node.id ?? null);
    }).catch(() => resolve(null));

    return () => {
      cancelled = true;
    };
  }, [documentId, environment, legacyVersionId, setSearchParams]);
}
