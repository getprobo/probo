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

import { dateTimeFormat } from "@probo/i18n";
import { Badge, Button, Card, Slack } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageSlackConnectionCardFragment$key } from "#/__generated__/core/CompliancePageSlackConnectionCardFragment.graphql";

const fragment = graphql`
  fragment CompliancePageSlackConnectionCardFragment on SlackConnection {
    id
    channel
    channelId
    createdAt
    canDelete: permission(action: "core:connector:delete")
  }
`;

interface Props {
  slackConnectionKey: CompliancePageSlackConnectionCardFragment$key;
  canConnect: boolean;
  buildConnectionUrl: (connectorId: string) => string;
  onDisconnect: (slackConnectionId: string) => void;
}

export function CompliancePageSlackConnectionCard(props: Props) {
  const { slackConnectionKey, canConnect, buildConnectionUrl, onDisconnect } = props;
  const { t, i18n } = useTranslation("organizations/compliance-page");

  const slackConnection = useFragment(fragment, slackConnectionKey);

  // A channel is only set once the incoming-webhook scope is granted, so its
  // absence means the connector isn't usable for the compliance page yet.
  const isConfigured = Boolean(slackConnection.channelId);

  return (
    <Card padded className="flex items-center gap-3">
      <div className="h-10 w-10 flex items-center justify-center bg-subtle rounded">
        <Slack className="h-6 w-6" />
      </div>
      <div className="mr-auto">
        <h3 className="text-base font-semibold">{t("slackConnectionCard.title")}</h3>
        <p className="text-sm text-txt-tertiary">
          {isConfigured
            ? (
                t(
                  slackConnection.channel
                    ? "slackConnectionCard.connectedWithChannel"
                    : "slackConnectionCard.connected",
                  {
                    date: dateTimeFormat(i18n.language, slackConnection.createdAt),
                    channel: slackConnection.channel,
                  },
                )
              )
            : (
                t("slackConnectionCard.notConfiguredDescription")
              )}
        </p>
      </div>
      <div className="flex items-center gap-2">
        {isConfigured
          ? (
              <Badge variant="success" size="md">
                {t("slackConnectionCard.status.connected")}
              </Badge>
            )
          : (
              <Badge variant="warning" size="md">
                {t("slackConnectionCard.status.notConfigured")}
              </Badge>
            )}
        {canConnect && (
          <Button variant="secondary" asChild>
            <a href={buildConnectionUrl(slackConnection.id)}>
              {isConfigured ? t("slackConnectionCard.actions.changeChannel") : t("slackConnectionCard.actions.connect")}
            </a>
          </Button>
        )}
        {slackConnection.canDelete && (
          <Button
            variant="secondary"
            onClick={() => onDisconnect(slackConnection.id)}
          >
            {t("slackConnectionCard.actions.disconnect")}
          </Button>
        )}
      </div>
    </Card>
  );
}
