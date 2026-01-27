import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentVersionsDropdownMenuQuery } from "#/__generated__/core/DocumentVersionsDropdownMenuQuery.graphql";

import { DocumentVersionsDropdownItem } from "./DocumentVersionsDropdownItem";

export const documentVersionsDropdownMenuQuery = graphql`
  query DocumentVersionsDropdownMenuQuery($documentId: ID!) {
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
      }
    }
  }
`;

export function DocumentVersionsDropdownMenu(props: {
  currentVersionId: string;
  queryRef: PreloadedQuery<DocumentVersionsDropdownMenuQuery>;
}) {
  const { currentVersionId, queryRef } = props;

  const { document } = usePreloadedQuery<DocumentVersionsDropdownMenuQuery>(
    documentVersionsDropdownMenuQuery,
    queryRef,
  );
  if (document.__typename !== "Document") {
    throw new Error("invalid type for node");
  }

  return (
    <>
      {document.versions.edges.map(({ node: version }) => (
        <DocumentVersionsDropdownItem
          key={version.id}
          fragmentRef={version}
          active={version.id === currentVersionId}
        />
      ))}
    </>
  );
}
