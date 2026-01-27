import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentVersionsDropdownMenuQuery } from "#/__generated__/core/DocumentVersionsDropdownMenuQuery.graphql";

import { DocumentVersionsDropdownItem } from "./DocumentVersionsDropdownItem";

export const documentVersionsDropdownMenuQuery = graphql`
  query DocumentVersionsDropdownMenuQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    document: node(id: $documentId) {
      __typename
      ... on Document {
        versions(first: 20) {
          edges {
            node {
              id
              ...DocumentVersionsDropdownItemFragment
            }
          }
        }
        # We use this on /documents/:documentId
        lastVersion: versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC }) @skip(if: $versionSpecified) {
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
}) {
  const { queryRef } = props;

  const { document, version } = usePreloadedQuery<DocumentVersionsDropdownMenuQuery>(
    documentVersionsDropdownMenuQuery,
    queryRef,
  );
  if (document.__typename !== "Document" || (version && version.__typename !== "DocumentVersion")) {
    throw new Error("invalid type for node");
  }

  const lastVersion = document.lastVersion?.edges[0].node;
  const currentVersion = lastVersion ?? version as NonNullable<typeof lastVersion | typeof version>;

  return (
    <>
      {document.versions.edges.map(({ node: version }) => (
        <DocumentVersionsDropdownItem
          key={version.id}
          fragmentRef={version}
          active={version.id === currentVersion.id}
        />
      ))}
    </>
  );
}
