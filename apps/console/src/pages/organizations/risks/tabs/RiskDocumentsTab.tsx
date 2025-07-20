import { graphql, useFragment } from "react-relay";
import { useOutletContext } from "react-router";
import { LinkedDocumentsCard } from "/components/documents/LinkedDocumentsCard";
import type { RiskDocumentsTabFragment$key } from "./__generated__/RiskDocumentsTabFragment.graphql";
import { useMutationWithIncrement } from "/hooks/useMutationWithIncrement.ts";

export const documentsFragment = graphql`
  fragment RiskDocumentsTabFragment on Risk {
    id
    documents(first: 100) @connection(key: "Risk__documents") {
      __id
      edges {
        node {
          id
          ...LinkedDocumentsCardFragment
        }
      }
    }
  }
`;

const attachDocumentMutation = graphql`
  mutation RiskDocumentsTabCreateMutation(
    $input: CreateRiskDocumentMappingInput!
    $connections: [ID!]!
  ) {
    createRiskDocumentMapping(input: $input) {
      documentEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedDocumentsCardFragment
        }
      }
    }
  }
`;

export const detachDocumentMutation = graphql`
  mutation RiskDocumentsTabDetachMutation(
    $input: DeleteRiskDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskDocumentMapping(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

export default function RiskDocumentsTab() {
  const { risk } = useOutletContext<{
    risk: RiskDocumentsTabFragment$key & { id: string };
  }>();
  const data = useFragment(documentsFragment, risk);
  const connectionId = data.documents.__id;
  const documents = data.documents?.edges?.map((edge) => edge.node) ?? [];

  const incrementOptions = {
    id: data.id,
    node: "documents(first:0)",
  };
  const [detachDocument, isDetaching] = useMutationWithIncrement(
    detachDocumentMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );
  const [attachDocument, isAttaching] = useMutationWithIncrement(
    attachDocumentMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedDocumentsCard
      disabled={isLoading}
      documents={documents}
      onAttach={attachDocument}
      onDetach={detachDocument}
      params={{ riskId: data.id }}
      connectionId={connectionId}
    />
  );
}
