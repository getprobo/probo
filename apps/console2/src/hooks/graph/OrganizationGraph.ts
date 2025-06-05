import { graphql } from "relay-runtime";

export const organizationViewQuery = graphql`
  query OrganizationGraph_ViewQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        id
        name
        ...SettingsPageFragment
      }
    }
  }
`;
