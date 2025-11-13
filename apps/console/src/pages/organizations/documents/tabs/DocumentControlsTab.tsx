import { LinkedControlsCard } from "/components/controls/LinkedControlsCard";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery, useRefetchableFragment } from "react-relay";
import type { DocumentControlsTabFragment$key } from "./__generated__/DocumentControlsTabFragment.graphql";
import type { DocumentControlsTabQuery } from "./__generated__/DocumentControlsTabQuery.graphql";
import { useMutationWithIncrement } from "/hooks/useMutationWithIncrement";

const documentControlsQuery = graphql`
  query DocumentControlsTabQuery($documentId: ID!) {
    node(id: $documentId) {
      ... on Document {
        ...DocumentControlsTabFragment
      }
    }
  }
`;

export const controlsFragment = graphql`
  fragment DocumentControlsTabFragment on Document
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "ControlOrder", defaultValue: null }
    filter: { type: "ControlFilter", defaultValue: null }
  )
  @refetchable(queryName: "DocumentControlsTabControlsQuery") {
    id
    controls(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "DocumentControlsTab_controls") {
      __id
      edges {
        node {
          id
          ...LinkedControlsCardFragment
        }
      }
    }
  }
`;

const detachControlMutation = graphql`
  mutation DocumentControlsTab_detachControlMutation(
    $input: DeleteControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlDocumentMapping(input: $input) {
      deletedControlId @deleteEdge(connections: $connections)
    }
  }
`;

const attachControlMutation = graphql`
  mutation DocumentControlsTab_attachControlMutation(
    $input: CreateControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    createControlDocumentMapping(input: $input) {
      controlEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedControlsCardFragment
        }
      }
    }
  }
`;

export default function DocumentControlsTab() {
  const { documentId } = useParams<{ documentId: string }>();
  const queryData = useLazyLoadQuery<DocumentControlsTabQuery>(
    documentControlsQuery,
    { documentId: documentId! }
  );
  const [data, refetch] = useRefetchableFragment(
    controlsFragment,
    queryData.node as DocumentControlsTabFragment$key
  );
  const controls = data.controls.edges.map((edge) => edge.node);
  const incrementOptions = {
    id: data.id,
    node: "controls(first:0)",
  };
  const [detachControl, isDetaching] = useMutationWithIncrement(
    detachControlMutation,
    {
      ...incrementOptions,
      value: -1,
      errorMessage: "Failed to unlink control",
    },
  );
  const [attachControl, isAttaching] = useMutationWithIncrement(
    attachControlMutation,
    {
      ...incrementOptions,
      value: 1,
      errorMessage: "Failed to link control",
    },
  );
  const isLoading = isDetaching || isAttaching;
  return (
    <LinkedControlsCard
      disabled={isLoading}
      controls={controls}
      params={{ documentId: data.id }}
      connectionId={data.controls.__id}
      onDetach={detachControl}
      onAttach={attachControl}
      refetch={refetch}
    />
  );
}
