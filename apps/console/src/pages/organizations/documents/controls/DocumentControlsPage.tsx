import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentControlsPageQuery } from "#/__generated__/core/DocumentControlsPageQuery.graphql";

import { DocumentControlList } from "./_components/DocumentControlList";

export const documentControlsPagePageQuery = graphql`
  query DocumentControlsPageQuery($documentId: ID!) {
    document: node(id: $documentId) {
      ...DocumentControlListFragment
    }
  }
`;

export function DocumentControlsPage(props: { queryRef: PreloadedQuery<DocumentControlsPageQuery> }) {
  const { queryRef } = props;

  const { document } = usePreloadedQuery<DocumentControlsPageQuery>(documentControlsPagePageQuery, queryRef);

  return <DocumentControlList fragmentRef={document} />;
}
