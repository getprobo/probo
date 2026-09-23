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

import { PlusMinusIcon, WebhooksLogoIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, useFragment, useRelayEnvironment } from "react-relay";
import { graphql } from "relay-runtime";

import type { WebhookSubscriptionListItem_updateMutation } from "#/__generated__/core/WebhookSubscriptionListItem_updateMutation.graphql";
import type { WebhookSubscriptionListItem_webhookSubscription$key } from "#/__generated__/core/WebhookSubscriptionListItem_webhookSubscription.graphql";
import type {
  WebhookEventStatus,
  WebhookSubscriptionListItemLastDeliveryQuery,
} from "#/__generated__/core/WebhookSubscriptionListItemLastDeliveryQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import type { WebhookEventTypeValue } from "../_lib/webhookEventTypes";
import { webhookSubscriptionListItem } from "../variants";

import { DeleteWebhookSubscriptionDialog } from "./DeleteWebhookSubscriptionDialog";
import { WebhookEventTypeBadges } from "./WebhookEventTypeBadges";
import { WebhookEventTypeSelectPopover } from "./WebhookEventTypeSelectPopover";

const webhookSubscriptionListItemFragment = graphql`
  fragment WebhookSubscriptionListItem_webhookSubscription on WebhookSubscription {
    id
    endpointUrl
    selectedEvents
    canUpdate: permission(action: "core:webhook-subscription:update")
    canDelete: permission(action: "core:webhook-subscription:delete")
  }
`;

const webhookSubscriptionListItemLastDeliveryQuery = graphql`
  query WebhookSubscriptionListItemLastDeliveryQuery($webhookSubscriptionId: ID!) {
    node(id: $webhookSubscriptionId) {
      __typename
      ... on WebhookSubscription {
        lastDeliveries: events(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              status
              createdAt
            }
          }
        }
      }
    }
  }
`;

function LastDeliveryStatusBadge({ status }: { status: WebhookEventStatus }) {
  const { t } = useTranslation();
  if (status === "SUCCEEDED") {
    return (
      <Badge variant="soft" color="green" size={1}>
        {t("webhooksSettingsPage.status.succeeded")}
      </Badge>
    );
  }
  if (status === "PENDING") {
    return (
      <Badge variant="soft" color="sky" size={1}>
        {t("webhooksSettingsPage.status.pending")}
      </Badge>
    );
  }
  return (
    <Badge variant="soft" color="red" size={1}>
      {t("webhooksSettingsPage.status.failed")}
    </Badge>
  );
}

function LastDeliveryActivity({ webhookSubscriptionId }: { webhookSubscriptionId: string }) {
  const { t, i18n } = useTranslation();
  const environment = useRelayEnvironment();
  const { activity } = webhookSubscriptionListItem();
  const [lastDelivery, setLastDelivery] = useState<{
    status: WebhookEventStatus;
    createdAt: string;
  } | null>();

  useEffect(() => {
    setLastDelivery(undefined);
    let cancelled = false;
    const subscription = fetchQuery<WebhookSubscriptionListItemLastDeliveryQuery>(
      environment,
      webhookSubscriptionListItemLastDeliveryQuery,
      { webhookSubscriptionId },
      { fetchPolicy: "network-only" },
    ).subscribe({
      next(data) {
        if (cancelled) {
          return;
        }
        if (data.node.__typename !== "WebhookSubscription") {
          setLastDelivery(null);
          return;
        }
        setLastDelivery(data.node.lastDeliveries.edges[0]?.node ?? null);
      },
      error() {
        if (!cancelled) {
          setLastDelivery(null);
        }
      },
    });
    return () => {
      cancelled = true;
      subscription.unsubscribe();
    };
  }, [environment, webhookSubscriptionId]);

  return (
    <div className={activity()}>
      {lastDelivery === undefined
        ? <TextSkeleton size={1} className="w-36" />
        : lastDelivery == null
          ? (
              <Text size={1} color="faint">
                {t("webhooksSettingsPage.lastDeliveryEmpty")}
              </Text>
            )
          : (
              <>
                <LastDeliveryStatusBadge status={lastDelivery.status} />
                <Text size={1} color="faint">
                  {dateTimeFormat(i18n.language, lastDelivery.createdAt)}
                </Text>
              </>
            )}
    </div>
  );
}

const updateWebhookSubscriptionMutation = graphql`
  mutation WebhookSubscriptionListItem_updateMutation(
    $input: UpdateWebhookSubscriptionInput!
  ) {
    updateWebhookSubscription(input: $input) {
      webhookSubscription {
        id
        selectedEvents
        updatedAt
      }
    }
  }
`;

interface WebhookSubscriptionListItemProps {
  webhookSubscriptionKey: WebhookSubscriptionListItem_webhookSubscription$key;
  onDeleted: () => void;
}

export function WebhookSubscriptionListItem({
  webhookSubscriptionKey,
  onDeleted,
}: WebhookSubscriptionListItemProps) {
  const { t } = useTranslation();
  const { card, header, lead, icon, endpoint, action, eventsSection, eventsHeading }
    = webhookSubscriptionListItem();
  const webhook = useFragment(webhookSubscriptionListItemFragment, webhookSubscriptionKey);
  const [updateWebhook, isUpdating] = useMutation<WebhookSubscriptionListItem_updateMutation>(
    updateWebhookSubscriptionMutation,
    {
      errorToast: t("webhooksSettingsPage.errors.update"),
    },
  );

  function handleToggleEvent(event: WebhookEventTypeValue) {
    const selected = webhook.selectedEvents.includes(event);
    const nextEvents = selected
      ? webhook.selectedEvents.filter(value => value !== event)
      : [...webhook.selectedEvents, event];
    if (nextEvents.length === 0) {
      return;
    }

    void updateWebhook({
      variables: {
        input: {
          id: webhook.id,
          selectedEvents: nextEvents,
        },
      },
    });
  }

  return (
    <Card variant="soft" size={2} interactive className={card()}>
      <div className={header()}>
        <div className={lead()}>
          <span className={icon()} aria-hidden>
            <WebhooksLogoIcon />
          </span>
          <Link to={webhook.id} underline={false} highContrast className={endpoint()}>
            <Code variant="ghost">{webhook.endpointUrl}</Code>
          </Link>
        </div>
        {webhook.canDelete && (
          <div className={action()}>
            <DeleteWebhookSubscriptionDialog
              webhookSubscriptionId={webhook.id}
              onDeleted={onDeleted}
            />
          </div>
        )}
      </div>
      <div className={eventsSection()}>
        <div className={eventsHeading()}>
          <Text size={2} weight="medium">
            {t("webhooksSettingsPage.subscribedCount", {
              count: webhook.selectedEvents.length,
            })}
          </Text>
          {webhook.canUpdate && (
            <WebhookEventTypeSelectPopover
              selectedEvents={webhook.selectedEvents}
              disabled={isUpdating}
              onToggle={handleToggleEvent}
            >
              <IconButton
                variant="surface"
                color="neutral"
                size={1}
                className={action()}
                aria-label={t("webhooksSettingsPage.editEvents")}
              >
                <PlusMinusIcon />
              </IconButton>
            </WebhookEventTypeSelectPopover>
          )}
        </div>
        <WebhookEventTypeBadges selectedEvents={webhook.selectedEvents} />
      </div>
      <LastDeliveryActivity webhookSubscriptionId={webhook.id} />
    </Card>
  );
}
