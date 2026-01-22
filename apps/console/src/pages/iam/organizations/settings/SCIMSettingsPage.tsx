import { useTranslate } from "@probo/i18n";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { SCIMSettingsPageQuery } from "/__generated__/iam/SCIMSettingsPageQuery.graphql";

import { SCIMConfiguration } from "./_components/SCIMConfiguration";
import { SCIMEventList } from "./_components/SCIMEventList";

export const scimSettingsPageQuery = graphql`
  query SCIMSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canCreateSCIMConfiguration: permission(
          action: "iam:scim-configuration:create"
        )
        canDeleteSCIMConfiguration: permission(
          action: "iam:scim-configuration:delete"
        )

        scimConfiguration {
          ...SCIMConfigurationFragment
          ...SCIMEventListFragment
        }
      }
    }
  }
`;

export function SCIMSettingsPage(props: {
  queryRef: PreloadedQuery<SCIMSettingsPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery(scimSettingsPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid node type");
  }

  return (
    <div className="space-y-8">
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("SCIM Provisioning")}</h2>
        <SCIMConfiguration
          fKey={organization.scimConfiguration ?? null}
          canCreate={organization.canCreateSCIMConfiguration}
          canDelete={organization.canDeleteSCIMConfiguration}
        />
      </div>

      {organization.scimConfiguration && (
        <div className="space-y-4">
          <h2 className="text-base font-medium">{__("SCIM Event History")}</h2>
          <SCIMEventList fKey={organization.scimConfiguration} />
        </div>
      )}
    </div>
  );
}
