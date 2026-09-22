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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { WebhookSubscriptionListItem_updateMutation } from "#/__generated__/core/WebhookSubscriptionListItem_updateMutation.graphql";
import type { WebhookSubscriptionListItem_webhookSubscription$key } from "#/__generated__/core/WebhookSubscriptionListItem_webhookSubscription.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import type { WebhookEventTypeValue } from "../_lib/webhookEventTypes";
import { webhookSubscriptionListItem } from "../variants";

import { DeleteWebhookSubscriptionDialog } from "./DeleteWebhookSubscriptionDialog";
import { WebhookEventsDialog } from "./WebhookEventsDialog";
import { WebhookEventTypeBadges } from "./WebhookEventTypeBadges";
import { WebhookEventTypeSelectPopover } from "./WebhookEventTypeSelectPopover";

const webhookSubscriptionListItemFragment = graphql`
  fragment WebhookSubscriptionListItem_webhookSubscription on WebhookSubscription {
    id
    endpointUrl
    selectedEvents
    canUpdate: permission(action: "core:webhook-subscription:update")
    canDelete: permission(action: "core:webhook-subscription:delete")
    events(first: 0) {
      totalCount
    }
  }
`;

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
  const { card, header, lead, icon, endpoint, eventsSection, eventsHeading, footer }
    = webhookSubscriptionListItem();
  const webhook = useFragment(webhookSubscriptionListItemFragment, webhookSubscriptionKey);
  const [viewingEvents, setViewingEvents] = useState(false);
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
    <Card variant="soft" size={2} className={card()}>
      <div className={header()}>
        <div className={lead()}>
          <span className={icon()} aria-hidden>
            <WebhooksLogoIcon />
          </span>
          <Code variant="ghost" className={endpoint()}>
            {webhook.endpointUrl}
          </Code>
        </div>
        {webhook.canDelete && (
          <DeleteWebhookSubscriptionDialog
            webhookSubscriptionId={webhook.id}
            onDeleted={onDeleted}
          />
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
                variant="outline"
                color="neutral"
                size={1}
                aria-label={t("webhooksSettingsPage.editEvents")}
              >
                <PlusMinusIcon />
              </IconButton>
            </WebhookEventTypeSelectPopover>
          )}
        </div>
        <WebhookEventTypeBadges selectedEvents={webhook.selectedEvents} />
      </div>
      <div className={footer()}>
        <Button
          variant="surface"
          color="neutral"
          onClick={() => {
            setViewingEvents(true);
          }}
        >
          {t("webhooksSettingsPage.eventsCount", { count: webhook.events.totalCount })}
        </Button>
      </div>
      {viewingEvents && (
        <WebhookEventsDialog
          webhookSubscriptionId={webhook.id}
          endpointUrl={webhook.endpointUrl}
          onClose={() => {
            setViewingEvents(false);
          }}
        />
      )}
    </Card>
  );
}
