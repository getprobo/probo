/*
const attachMeasureMutation = graphql`
  mutation MeasureLinkDialogCreateMutation(
    $input: CreateRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    createRiskMeasureMapping(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          category
          state
        }
      }
    }
  }
`;

export const detachMeasureMutation = graphql`
  mutation MeasureLinkDialogDetachMutation(
    $input: DeleteRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskMeasureMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;
*/
