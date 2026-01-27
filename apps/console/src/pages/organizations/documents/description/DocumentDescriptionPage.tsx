import { Markdown } from "@probo/ui";
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

  const lastVersion = document.lastVersion?.edges[0].node;
  const currentVersion = lastVersion ?? version as NonNullable<typeof lastVersion | typeof version>;

  return (
    <div>
      <Markdown content={currentVersion.content} />
    </div>
  );
}
