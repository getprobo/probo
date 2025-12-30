import { graphql } from "relay-runtime";

export const domainSettingsPagePermissionsQuery = graphql`
  query domainSettingsPage_permissionsQuery($organizationId: ID!) {
    viewer @required(action: THROW) {
      canCreateCustomDomain: permission(
        action: "core:custom-domain:create"
        id: $organizationId
      )
      canDeleteCustomDomain: permission(
        action: "core:custom-domain:delete"
        id: $organizationId
      )
    }
  }
`;
