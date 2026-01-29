import { useTranslate } from "@probo/i18n";
import { graphql, useFragment } from "react-relay";

import type { ConnectorListFragment$key } from "#/__generated__/iam/ConnectorListFragment.graphql";

import { GoogleWorkspaceConnector } from "./GoogleWorkspaceConnector";

const connectorListFragment = graphql`
  fragment ConnectorListFragment on Organization {
    scimConfiguration {
      ...GoogleWorkspaceConnectorFragment
    }
  }
`;

export function ConnectorList(props: { fKey: ConnectorListFragment$key }) {
  const { fKey } = props;
  const data = useFragment<ConnectorListFragment$key>(connectorListFragment, fKey);
  const { __ } = useTranslate();

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{__("Identity Provider")}</h2>
      <p className="text-sm text-txt-secondary">
        {__(
          "Connect your identity provider to automatically sync users to your organization. Once connected, you don't need to configure SCIM manually.",
        )}
      </p>
      <GoogleWorkspaceConnector fKey={data.scimConfiguration ?? null} />
    </div>
  );
}
