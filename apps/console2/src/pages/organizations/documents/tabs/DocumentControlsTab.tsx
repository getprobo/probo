import { LinkedControlsCard } from "/components/controls/LinkedControlsCard";
import { useOutletContext } from "react-router";
import type { DocumentDetailPageDocumentFragment$data } from "../__generated__/DocumentDetailPageDocumentFragment.graphql";
import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";

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

export default function DocumentControlsTab() {
  const { document } = useOutletContext<{
    document: DocumentDetailPageDocumentFragment$data;
  }>();
  const controls = document.controls.edges.map((edge) => edge.node);
  const [detachControl] = useMutation(detachControlMutation);
  console.log(controls.map((c) => c.id));
  return (
    <LinkedControlsCard
      controls={controls}
      params={{ documentId: document.id }}
      connectionId={document.controls.__id}
      onDetach={detachControl}
    />
  );
}
