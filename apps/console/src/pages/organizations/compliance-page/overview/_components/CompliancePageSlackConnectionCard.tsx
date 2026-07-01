// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Card, Slack } from "@probo/ui";
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
  const { __, dateTimeFormat } = useTranslate();

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
        <h3 className="text-base font-semibold">Slack</h3>
        <p className="text-sm text-txt-tertiary">
          {isConfigured
            ? (
                <>
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
                </>
              )
            : (
                __("Slack is connected, but no channel is set up for the compliance page yet.")
              )}
        </p>
      </div>
      <div className="flex items-center gap-2">
        {isConfigured
          ? (
              <Badge variant="success" size="md">
                {__("Connected")}
              </Badge>
            )
          : (
              <Badge variant="warning" size="md">
                {__("Channel not configured")}
              </Badge>
            )}
        {canConnect && (
          <Button variant="secondary" asChild>
            <a href={buildConnectionUrl(slackConnection.id)}>
              {isConfigured ? __("Change channel") : __("Connect")}
            </a>
          </Button>
        )}
        {slackConnection.canDelete && (
          <Button
            variant="secondary"
            onClick={() => onDisconnect(slackConnection.id)}
          >
            {__("Disconnect")}
          </Button>
        )}
      </div>
    </Card>
  );
}
