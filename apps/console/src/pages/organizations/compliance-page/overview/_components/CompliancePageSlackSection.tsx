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

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Card, Slack, useConfirm } from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageSlackSectionDeleteMutation } from "#/__generated__/core/CompliancePageSlackSectionDeleteMutation.graphql";
import type { CompliancePageSlackSectionFragment$key } from "#/__generated__/core/CompliancePageSlackSectionFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment CompliancePageSlackSectionFragment on Organization {
    canConnectSlack: permission(action: "core:connector:initiate")
    slackOAuth2Scopes
    slackConnections(first: 100) {
      __id
      edges {
        node {
          id
          channel
          createdAt
          canDelete: permission(action: "core:connector:delete")
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
  const { __, dateTimeFormat } = useTranslate();
  const confirm = useConfirm();

  const organization = useFragment<CompliancePageSlackSectionFragment$key>(fragment, fragmentRef);
  const [deleteSlackConnection] = useMutation<CompliancePageSlackSectionDeleteMutation>(deleteMutation);

  const connectionId = organization.slackConnections.__id;

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
        title: __("Disconnect Slack"),
        message: __("Are you sure you want to disconnect this Slack channel? This action cannot be undone."),
        label: __("Disconnect"),
        variant: "danger",
      },
    );
  };

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{__("Integrations")}</h2>
      <div className="space-y-2">
        {organization.slackConnections.edges.map(({ node: slackConnection }) => (
          <Card
            key={slackConnection.id}
            padded
            className="flex items-center gap-3"
          >
            <div className="h-10 w-10 flex items-center justify-center bg-subtle rounded">
              <Slack className="h-6 w-6" />
            </div>
            <div className="mr-auto">
              <h3 className="text-base font-semibold">Slack</h3>
              <p className="text-sm text-txt-tertiary">
                {sprintf(
                  __("Connected on %s"),
                  dateTimeFormat(slackConnection.createdAt),
                )}
                {slackConnection.channel && (
                  <>
                    {" • "}
                    {sprintf(__("Channel: %s"), slackConnection.channel)}
                  </>
                )}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant="success" size="md">
                {__("Connected")}
              </Badge>
              {slackConnection.canDelete && (
                <Button
                  variant="secondary"
                  onClick={() => handleDisconnect(slackConnection.id)}
                >
                  {__("Disconnect")}
                </Button>
              )}
            </div>
          </Card>
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
              <h3 className="text-base font-semibold">Slack</h3>
              <p className="text-sm text-txt-tertiary">
                {__("Manage your compliance page access with slack")}
              </p>
            </div>
            <Button variant="secondary" asChild>
              <a href={getSlackConnectionUrl(organizationId, organization.slackOAuth2Scopes)}>
                {__("Connect")}
              </a>
            </Button>
          </Card>
        )}
      </div>
    </div>
  );
}

function getSlackConnectionUrl(organizationId: string, scopes: readonly string[]): string {
  const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/connectors/initiate", baseUrl);
  url.searchParams.append("organization_id", organizationId);
  url.searchParams.append("provider", "SLACK");
  for (const scope of scopes) {
    url.searchParams.append("scope", scope);
  }
  const redirectUrl = `/organizations/${organizationId}/compliance-page`;
  url.searchParams.append("continue", redirectUrl);
  return url.toString();
}
