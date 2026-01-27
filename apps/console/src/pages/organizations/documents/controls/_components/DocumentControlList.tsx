import type { ComponentProps } from "react";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentControlListFragment$key } from "#/__generated__/core/DocumentControlListFragment.graphql";
import type { DocumentControlListQuery } from "#/__generated__/core/DocumentControlListQuery.graphql";
import { LinkedControlsCard } from "#/components/controls/LinkedControlsCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

const fragment = graphql`
  fragment DocumentControlListFragment on Document
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "ControlOrder", defaultValue: null }
    filter: { type: "ControlFilter", defaultValue: null }
  )
  @refetchable(queryName: "DocumentControlListQuery") {
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
  mutation DocumentControlList_detachControlMutation(
    $input: DeleteControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlDocumentMapping(input: $input) {
      deletedControlId @deleteEdge(connections: $connections)
    }
  }
`;

const attachControlMutation = graphql`
  mutation DocumentControlList_attachControlMutation(
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

export function DocumentControlList(props: { fragmentRef: DocumentControlListFragment$key }) {
  const { fragmentRef } = props;

  const [document, refetch] = useRefetchableFragment<DocumentControlListQuery, DocumentControlListFragment$key>(
    fragment,
    fragmentRef,
  );
  const incrementOptions = {
    id: document.id,
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
      controls={document.controls.edges.map(({ node }) => node)}
      params={{ documentId: document.id }}
      connectionId={document.controls.__id}
      onDetach={detachControl}
      onAttach={attachControl}
      refetch={refetch as ComponentProps<typeof LinkedControlsCard>["refetch"]}
    />
  );
}
