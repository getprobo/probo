import { graphql } from "relay-runtime";

export const evidenceFileQuery = graphql`
  query EvidenceGraphFileQuery($evidenceId: ID!) {
    node(id: $evidenceId) {
      ... on Evidence {
        id
        file {
            mimeType
            fileName
            size
            downloadUrl
        }
      }
    }
  }
`;
