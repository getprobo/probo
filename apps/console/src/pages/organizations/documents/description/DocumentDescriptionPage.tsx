// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { RichEditor } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentDescriptionPageQuery } from "#/__generated__/core/DocumentDescriptionPageQuery.graphql";

export const documentDescriptionPageQuery = graphql`
  query DocumentDescriptionPageQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    # We use this on /documents/:documentId/versions/:versionId/description
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ... on DocumentVersion {
        content
      }
    }
    document: node(id: $documentId) {
      __typename
      ... on Document {
        # We use this on /documents/:documentId/description
        lastVersion: versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC }) @skip(if: $versionSpecified) {
          edges {
            node {
              content
            }
          }
        }
      }
    }
  }
`;

export function DocumentDescriptionPage(props: { queryRef: PreloadedQuery<DocumentDescriptionPageQuery> }) {
  const { queryRef } = props;

  const { document, version } = usePreloadedQuery<DocumentDescriptionPageQuery>(
    documentDescriptionPageQuery,
    queryRef,
  );
  if (document.__typename !== "Document" || (version && version.__typename !== "DocumentVersion")) {
    throw new Error("invalid type for node");
  }

  // const lastVersion = document.lastVersion?.edges[0].node;
  // const currentVersion = lastVersion ?? version as NonNullable<typeof lastVersion | typeof version>;

  return (
    <div>
      <RichEditor />
      {/* <Markdown content={currentVersion.content} /> */}
    </div>
  );
}
