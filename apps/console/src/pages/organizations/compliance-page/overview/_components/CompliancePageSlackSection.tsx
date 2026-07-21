// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Button, Card, Slack, useConfirm } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageSlackSectionDeleteMutation } from "#/__generated__/core/CompliancePageSlackSectionDeleteMutation.graphql";
import type { CompliancePageSlackSectionFragment$key } from "#/__generated__/core/CompliancePageSlackSectionFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CompliancePageSlackConnectionCard } from "./CompliancePageSlackConnectionCard";

const fragment = graphql`
  fragment CompliancePageSlackSectionFragment on Organization {
    canConnectSlack: permission(action: "core:connector:initiate")
    slackOAuth2Scopes
    slackConnections(first: 100) {
      __id
      edges {
        node {
          id
          ...CompliancePageSlackConnectionCardFragment
        }
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation CompliancePageSlackSectionDeleteMutation(
    $input: DeleteSlackConnectionInput!
    $connections: [ID!]!
  ) {
    deleteSlackConnection(input: $input) {
      deletedSlackConnectionId @deleteEdge(connections: $connections)
    }
  }
`;

export function CompliancePageSlackSection(props: { fragmentRef: CompliancePageSlackSectionFragment$key }) {
  const { fragmentRef } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-page");
  const confirm = useConfirm();

  const organization = useFragment<CompliancePageSlackSectionFragment$key>(fragment, fragmentRef);
  const [deleteSlackConnection] = useMutation<CompliancePageSlackSectionDeleteMutation>(deleteMutation);

  const connectionId = organization.slackConnections.__id;
  const scopes = organization.slackOAuth2Scopes;

  const handleDisconnect = (slackConnectionId: string) => {
    confirm(
      () =>
        new Promise<void>((resolve, reject) => {
          deleteSlackConnection({
            variables: {
              connections: [connectionId],
              input: {
                slackConnectionId,
              },
            },
            onCompleted: () => resolve(),
            onError: error => reject(error),
          });
        }),
      {
        title: t("slackSection.disconnect.title"),
        message: t("slackSection.disconnect.description"),
        label: t("slackSection.actions.disconnect"),
        variant: "danger",
      },
    );
  };

  // Passing connector_id reconnects in place (union of scopes), letting users
  // pick a channel without disconnecting first.
  const buildConnectionUrl = (connectorId?: string) =>
    getSlackConnectionUrl(organizationId, scopes, connectorId);

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{t("slackSection.title")}</h2>
      <div className="space-y-2">
        {organization.slackConnections.edges.map(({ node: slackConnection }) => (
          <CompliancePageSlackConnectionCard
            key={slackConnection.id}
            slackConnectionKey={slackConnection}
            canConnect={organization.canConnectSlack}
            buildConnectionUrl={connectorId => buildConnectionUrl(connectorId)}
            onDisconnect={handleDisconnect}
          />
        ))}
        {organization.canConnectSlack && organization.slackConnections.edges.length === 0 && (
          <Card
            padded
            className="flex items-center gap-3"
          >
            <div className="h-10 w-10 flex items-center justify-center bg-subtle rounded">
              <Slack className="h-6 w-6" />
            </div>
            <div className="mr-auto">
              <h3 className="text-base font-semibold">{t("slackConnectionCard.title")}</h3>
              <p className="text-sm text-txt-tertiary">
                {t("slackSection.emptyDescription")}
              </p>
            </div>
            <Button variant="secondary" asChild>
              <a href={buildConnectionUrl()}>
                {t("slackSection.actions.connect")}
              </a>
            </Button>
          </Card>
        )}
      </div>
    </div>
  );
}

function getSlackConnectionUrl(
  organizationId: string,
  scopes: readonly string[],
  connectorId?: string,
): string {
  const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/connectors/initiate", baseUrl);
  url.searchParams.append("organization_id", organizationId);
  url.searchParams.append("provider", "SLACK");
  for (const scope of scopes) {
    url.searchParams.append("scope", scope);
  }
  if (connectorId) {
    url.searchParams.append("connector_id", connectorId);
  }
  const redirectUrl = `/organizations/${organizationId}/compliance-page`;
  url.searchParams.append("continue", redirectUrl);
  return url.toString();
}
