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

import { Form } from "@base-ui/react/form";
import { CaretLeftIcon, PlusMinusIcon } from "@phosphor-icons/react";
import { toFieldErrors } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { WebhookSubscriptionDetailPage_updateMutation } from "#/__generated__/core/WebhookSubscriptionDetailPage_updateMutation.graphql";
import type { WebhookSubscriptionDetailPage_webhookSubscription$key } from "#/__generated__/core/WebhookSubscriptionDetailPage_webhookSubscription.graphql";
import type { WebhookSubscriptionDetailPageQuery } from "#/__generated__/core/WebhookSubscriptionDetailPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";

import { DeleteWebhookSubscriptionDialog } from "./_components/DeleteWebhookSubscriptionDialog";
import { WebhookEventTypeBadges } from "./_components/WebhookEventTypeBadges";
import { WebhookEventTypeSelectPopover } from "./_components/WebhookEventTypeSelectPopover";
import { WebhookSigningSecretField } from "./_components/WebhookSigningSecretField";
import { WebhookSubscriptionEventList } from "./_components/WebhookSubscriptionEventList";
import type { WebhookEventTypeValue } from "./_lib/webhookEventTypes";
import { webhookSubscriptionDetailPage } from "./variants";

export const webhookSubscriptionDetailPageFragment = graphql`
  fragment WebhookSubscriptionDetailPage_webhookSubscription on WebhookSubscription
  @argumentDefinitions(filter: { type: "WebhookEventFilter", defaultValue: null }) {
    id
    endpointUrl
    selectedEvents
    canUpdate: permission(action: "core:webhook-subscription:update")
    canDelete: permission(action: "core:webhook-subscription:delete")
    organization {
      id
    }
    ...WebhookSubscriptionEventList_webhookSubscription @arguments(filter: $filter)
  }
`;

export const webhookSubscriptionDetailPageQuery = graphql`
  query WebhookSubscriptionDetailPageQuery(
    $webhookSubscriptionId: ID!
    $filter: WebhookEventFilter
  ) {
    node(id: $webhookSubscriptionId) {
      __typename
      ... on WebhookSubscription {
        ...WebhookSubscriptionDetailPage_webhookSubscription @arguments(filter: $filter)
      }
    }
  }
`;

const updateWebhookSubscriptionMutation = graphql`
  mutation WebhookSubscriptionDetailPage_updateMutation(
    $input: UpdateWebhookSubscriptionInput!
  ) {
    updateWebhookSubscription(input: $input) {
      webhookSubscription {
        id
        endpointUrl
        selectedEvents
        updatedAt
      }
    }
  }
`;

function endpointUrlError(value: string, required: string, invalid: string, https: string) {
  const trimmed = value.trim();
  if (trimmed === "") {
    return required;
  }

  try {
    const url = new URL(trimmed);
    if (url.protocol !== "https:") {
      return https;
    }
  } catch {
    return invalid;
  }

  return undefined;
}

function clearFieldError(errors: Record<string, string>, field: string) {
  if (errors[field] == null) {
    return errors;
  }

  const next = { ...errors };
  delete next[field];
  return next;
}

interface WebhookSubscriptionDetailPageProps {
  queryRef: PreloadedQuery<WebhookSubscriptionDetailPageQuery>;
}

export function WebhookSubscriptionDetailPage({
  queryRef,
}: WebhookSubscriptionDetailPageProps) {
  const data = usePreloadedQuery<WebhookSubscriptionDetailPageQuery>(
    webhookSubscriptionDetailPageQuery,
    queryRef,
  );
  const { t } = useTranslation();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const webhookKey = data.node?.__typename === "WebhookSubscription" ? data.node : null;
  const webhook = useFragment<WebhookSubscriptionDetailPage_webhookSubscription$key>(
    webhookSubscriptionDetailPageFragment,
    webhookKey,
  );

  const title = webhook?.endpointUrl !== "" && webhook?.endpointUrl != null
    ? webhook.endpointUrl
    : t("nav.webhooks");
  usePageTitle(title);

  const [draftUrl, setDraftUrl] = useState(webhook?.endpointUrl ?? "");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [updateWebhook, isUpdating] = useMutation<WebhookSubscriptionDetailPage_updateMutation>(
    updateWebhookSubscriptionMutation,
    {
      successMessage: t("webhooksSettingsPage.messages.updated"),
      errorToast: t("webhooksSettingsPage.errors.update"),
    },
  );

  if (webhook == null || webhook.organization?.id !== organizationId) {
    throw new NotFoundError(t("webhooksSettingsPage.notFound"));
  }

  const subscription = webhook;
  const {
    root,
    back,
    header,
    intro,
    title: titleClass,
    settings,
    endpointRow,
    endpointField,
    eventsSection,
    eventsHeading,
    history,
  } = webhookSubscriptionDetailPage();

  const listPath = `/organizations/${organizationId}/settings/webhooks`;
  const isEndpointDirty = draftUrl.trim() !== subscription.endpointUrl;

  function handleSaveEndpoint() {
    const urlError = endpointUrlError(
      draftUrl,
      t("webhooksSettingsPage.validation.endpointUrlRequired"),
      t("webhooksSettingsPage.validation.invalidUrl"),
      t("webhooksSettingsPage.validation.httpsRequired"),
    );
    if (urlError != null) {
      setErrors({ endpointUrl: urlError });
      return;
    }

    void updateWebhook({
      variables: {
        input: {
          id: subscription.id,
          endpointUrl: draftUrl.trim(),
        },
      },
      onCompleted(_response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(fieldErrors);
        }
      },
    }).catch(() => {
      // Field errors are mapped in onCompleted; other failures toast.
    });
  }

  function handleToggleEvent(event: WebhookEventTypeValue) {
    const selected = subscription.selectedEvents.includes(event);
    const nextEvents = selected
      ? subscription.selectedEvents.filter(value => value !== event)
      : [...subscription.selectedEvents, event];
    if (nextEvents.length === 0) {
      return;
    }

    void updateWebhook({
      variables: {
        input: {
          id: subscription.id,
          selectedEvents: nextEvents,
        },
      },
    });
  }

  return (
    <div className={root()}>
      <Link
        to={listPath}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("webhooksSettingsPage.actions.back")}
      </Link>
      <div className={header()}>
        <div className={intro()}>
          <Heading
            level={1}
            size={6}
            weight="medium"
            highContrast
            className={titleClass()}
          >
            {title}
          </Heading>
        </div>
        {subscription.canDelete && (
          <DeleteWebhookSubscriptionDialog
            webhookSubscriptionId={subscription.id}
            trigger="button"
            onDeleted={() => {
              void navigate(listPath);
            }}
          />
        )}
      </div>
      <Card variant="soft" size={2}>
        <div className={settings()}>
          {subscription.canUpdate
            ? (
                <Form errors={errors} onFormSubmit={handleSaveEndpoint}>
                  <div className={endpointRow()}>
                    <Field
                      className={endpointField()}
                      label={t("webhooksSettingsPage.fields.endpointUrl")}
                      error={errors.endpointUrl}
                    >
                      <TextField
                        name="endpointUrl"
                        type="url"
                        required
                        value={draftUrl}
                        placeholder={t("webhooksSettingsPage.placeholders.endpointUrl")}
                        onValueChange={(value) => {
                          setDraftUrl(value);
                          setErrors(current => clearFieldError(current, "endpointUrl"));
                        }}
                      />
                    </Field>
                    {isEndpointDirty && (
                      <Button
                        type="submit"
                        variant="solid"
                        color="neutral"
                        highContrast
                        loading={isUpdating}
                      >
                        {t("webhooksSettingsPage.actions.save")}
                      </Button>
                    )}
                  </div>
                </Form>
              )
            : (
                <div>
                  <Text size={2} weight="medium">
                    {t("webhooksSettingsPage.fields.endpointUrl")}
                  </Text>
                  <Code variant="ghost">{subscription.endpointUrl}</Code>
                </div>
              )}
          <WebhookSigningSecretField webhookSubscriptionId={subscription.id} />
          <div className={eventsSection()}>
            <div className={eventsHeading()}>
              <Text size={2} weight="medium">
                {t("webhooksSettingsPage.subscribedCount", {
                  count: subscription.selectedEvents.length,
                })}
              </Text>
              {subscription.canUpdate && (
                <WebhookEventTypeSelectPopover
                  selectedEvents={subscription.selectedEvents}
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
            <WebhookEventTypeBadges selectedEvents={subscription.selectedEvents} />
          </div>
        </div>
      </Card>
      <div className={history()}>
        <WebhookSubscriptionEventList webhookSubscriptionKey={subscription} />
      </div>
    </div>
  );
}
