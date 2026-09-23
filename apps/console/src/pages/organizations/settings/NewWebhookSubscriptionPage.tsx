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
import { CaretLeftIcon } from "@phosphor-icons/react";
import { toFieldErrors } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Combobox } from "@probo/ui/src/v2/Combobox/Combobox";
import { ComboboxChip } from "@probo/ui/src/v2/Combobox/ComboboxChip";
import { ComboboxChipRemove } from "@probo/ui/src/v2/Combobox/ComboboxChipRemove";
import { ComboboxChips } from "@probo/ui/src/v2/Combobox/ComboboxChips";
import { ComboboxEmpty } from "@probo/ui/src/v2/Combobox/ComboboxEmpty";
import { ComboboxInput } from "@probo/ui/src/v2/Combobox/ComboboxInput";
import { ComboboxInputGroup } from "@probo/ui/src/v2/Combobox/ComboboxInputGroup";
import { ComboboxItem } from "@probo/ui/src/v2/Combobox/ComboboxItem";
import { ComboboxList } from "@probo/ui/src/v2/Combobox/ComboboxList";
import { ComboboxPopup } from "@probo/ui/src/v2/Combobox/ComboboxPopup";
import { ComboboxValue } from "@probo/ui/src/v2/Combobox/ComboboxValue";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { NewWebhookSubscriptionPageMutation } from "#/__generated__/core/NewWebhookSubscriptionPageMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import type { WebhookEventType } from "./_lib/webhookEventTypes";
import { WEBHOOK_EVENT_TYPES } from "./_lib/webhookEventTypes";
import { clearFieldError, endpointUrlError } from "./_lib/webhookSubscriptionForm";
import { newWebhookSubscriptionPage } from "./variants";

const createWebhookSubscriptionMutation = graphql`
  mutation NewWebhookSubscriptionPageMutation(
    $input: CreateWebhookSubscriptionInput!
    $connections: [ID!]!
  ) {
    createWebhookSubscription(input: $input) {
      webhookSubscriptionEdge @prependEdge(connections: $connections) {
        node {
          id
          ...WebhookSubscriptionListItem_webhookSubscription
        }
      }
    }
  }
`;

function NewWebhookSubscriptionPageInner() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const title = t("webhooksSettingsPage.dialogs.addTitle");
  usePageTitle(title);

  const { root, back, intro, form, actions } = newWebhookSubscriptionPage();

  const [endpointUrl, setEndpointUrl] = useState("");
  const [selected, setSelected] = useState<WebhookEventType[]>([]);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const [createWebhook, isCreating] = useMutation<NewWebhookSubscriptionPageMutation>(
    createWebhookSubscriptionMutation,
    {
      successMessage: t("webhooksSettingsPage.messages.created"),
      errorToast: t("webhooksSettingsPage.errors.create"),
    },
  );

  function handleSubmit() {
    const nextErrors: Record<string, string> = {};
    const urlError = endpointUrlError(
      endpointUrl,
      t("webhooksSettingsPage.validation.endpointUrlRequired"),
      t("webhooksSettingsPage.validation.invalidUrl"),
      t("webhooksSettingsPage.validation.httpsRequired"),
    );
    if (urlError != null) {
      nextErrors.endpointUrl = urlError;
    }
    if (selected.length === 0) {
      nextErrors.selectedEvents = t("webhooksSettingsPage.validation.eventRequired");
    }
    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      return;
    }

    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      "WebhookSubscriptionList_webhookSubscriptions",
    );

    void createWebhook({
      variables: {
        input: {
          organizationId,
          endpointUrl: endpointUrl.trim(),
          selectedEvents: selected.map(event => event.value),
        },
        connections: [connectionId],
      },
      onCompleted(response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(fieldErrors);
          return;
        }
        if (response.createWebhookSubscription.webhookSubscriptionEdge.node.id !== "") {
          void navigate(`/organizations/${organizationId}/settings/webhooks`);
        }
      },
    }).catch(() => {
      // Field errors are mapped in onCompleted; other failures toast.
    });
  }

  return (
    <div className={root()}>
      <Link
        to={`/organizations/${organizationId}/settings/webhooks`}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("webhooksSettingsPage.actions.back")}
      </Link>
      <div className={intro()}>
        <Heading level={1} size={6} weight="medium" highContrast>
          {title}
        </Heading>
        <Text size={2} color="faint">
          {t("webhooksSettingsPage.newDescription")}
        </Text>
      </div>
      <Card variant="soft" size={2}>
        <Form className={form()} errors={errors} onFormSubmit={handleSubmit}>
          <Field
            label={t("webhooksSettingsPage.fields.endpointUrl")}
            error={errors.endpointUrl}
          >
            <TextField
              name="endpointUrl"
              type="url"
              required
              value={endpointUrl}
              placeholder={t("webhooksSettingsPage.placeholders.endpointUrl")}
              onValueChange={(value) => {
                setEndpointUrl(value);
                setErrors(current => clearFieldError(current, "endpointUrl"));
              }}
            />
          </Field>
          <Field
            label={t("webhooksSettingsPage.fields.events")}
            error={errors.selectedEvents}
          >
            <Combobox
              multiple
              items={WEBHOOK_EVENT_TYPES}
              value={selected}
              onValueChange={(value) => {
                setSelected(value);
                setErrors(current => clearFieldError(current, "selectedEvents"));
              }}
              itemToStringLabel={item => item.label}
              isItemEqualToValue={(a, b) => a.value === b.value}
              disabled={isCreating}
            >
              <ComboboxInputGroup>
                <ComboboxChips>
                  <ComboboxValue<WebhookEventType[]>>
                    {(value) => {
                      const events = value ?? [];
                      return (
                        <>
                          {events.map(event => (
                            <ComboboxChip key={event.value} aria-label={event.label}>
                              {event.label}
                              <ComboboxChipRemove
                                aria-label={t("webhooksSettingsPage.removeEvent", { event: event.label })}
                              />
                            </ComboboxChip>
                          ))}
                          <ComboboxInput
                            placeholder={
                              events.length === 0
                                ? t("webhooksSettingsPage.searchEvents")
                                : undefined
                            }
                          />
                        </>
                      );
                    }}
                  </ComboboxValue>
                </ComboboxChips>
              </ComboboxInputGroup>
              <ComboboxPopup>
                <ComboboxEmpty>{t("webhooksSettingsPage.emptySearch")}</ComboboxEmpty>
                <ComboboxList<WebhookEventType>>
                  {event => (
                    <ComboboxItem key={event.value} value={event}>{event.label}</ComboboxItem>
                  )}
                </ComboboxList>
              </ComboboxPopup>
            </Combobox>
          </Field>
          <div className={actions()}>
            <Button
              type="submit"
              variant="solid"
              color="neutral"
              highContrast
              loading={isCreating}
            >
              {t("webhooksSettingsPage.actions.create")}
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  );
}

export default function NewWebhookSubscriptionPage() {
  return (
    <CoreRelayProvider>
      <NewWebhookSubscriptionPageInner />
    </CoreRelayProvider>
  );
}
