import { graphql } from "relay-runtime";

export const tasksQuery = graphql`
  query TaskGraphQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        ...TasksPageFragment
      }
    }
  }
`;
