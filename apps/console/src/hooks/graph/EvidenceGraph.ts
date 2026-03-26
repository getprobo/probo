import { graphql } from "relay-runtime";

/* eslint-disable relay/unused-fields */

export const evidenceFileQuery = graphql`
  query EvidenceGraphFileQuery($evidenceId: ID!) {
    node(id: $evidenceId) {
      ... on Evidence {
        id
        description
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
