import { graphql } from "relay-runtime";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

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
