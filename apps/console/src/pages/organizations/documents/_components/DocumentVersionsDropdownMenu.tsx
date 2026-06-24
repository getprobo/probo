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

import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentVersionsDropdownMenuQuery } from "#/__generated__/core/DocumentVersionsDropdownMenuQuery.graphql";

import { DocumentVersionsDropdownItem } from "./DocumentVersionsDropdownItem";

export const documentVersionsDropdownMenuQuery = graphql`
  query DocumentVersionsDropdownMenuQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    document: node(id: $documentId) {
      __typename
      ... on Document {
        versions(first: 20) @connection(key: "DocumentversionsDropdownMenu_versions") {
          edges {
            node {
              id
              ...DocumentVersionsDropdownItemFragment
            }
          }
        }
        # We use this on /documents/:documentId
        lastVersion: versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC })
        @skip(if: $versionSpecified)
        @connection(key: "DocumentversionsDropdownMenu_lastVersion" filters: ["orderBy"]) {
          edges {
            node {
              id
            }
          }
        }
      }
    }
    # We use this on /documents/:documentId/versions/:versionId
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ... on DocumentVersion {
        id
      }
    }
  }
`;

export function DocumentVersionsDropdownMenu(props: {
  queryRef: PreloadedQuery<DocumentVersionsDropdownMenuQuery>;
  currentTab: string | undefined;
}) {
  const { queryRef, currentTab } = props;

  const { document, version } = usePreloadedQuery<DocumentVersionsDropdownMenuQuery>(
    documentVersionsDropdownMenuQuery,
    queryRef,
  );
  if (document.__typename !== "Document" || (version && version.__typename !== "DocumentVersion")) {
    throw new Error("invalid type for node");
  }

  const lastVersion = document.lastVersion?.edges[0].node;
  const currentVersion = lastVersion ?? version;

  return (
    <>
      {document.versions.edges.map(({ node: version }) => (
        <DocumentVersionsDropdownItem
          key={version.id}
          fragmentRef={version}
          active={version.id === currentVersion?.id}
          currentTab={currentTab}
        />
      ))}
    </>
  );
}
