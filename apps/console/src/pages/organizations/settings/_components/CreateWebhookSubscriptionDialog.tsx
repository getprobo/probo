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

import { PlusIcon } from "@phosphor-icons/react";
import {
  Button as LegacyButton,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import type { TFunction } from "i18next";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { CreateWebhookSubscriptionDialogMutation } from "#/__generated__/core/CreateWebhookSubscriptionDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { z } from "#/lib/zod";

import {
  WEBHOOK_EVENT_TYPES,
  type WebhookEventTypeValue,
} from "../_lib/webhookEventTypes";

const createWebhookSubscriptionMutation = graphql`
  mutation CreateWebhookSubscriptionDialogMutation(
    $input: CreateWebhookSubscriptionInput!
  ) {
    createWebhookSubscription(input: $input) {
      webhookSubscriptionEdge {
        node {
          id
        }
      }
    }
  }
`;

const WEBHOOK_EVENT_VALUES = WEBHOOK_EVENT_TYPES.map(event => event.value) as [
  WebhookEventTypeValue,
  ...WebhookEventTypeValue[],
];

const createWebhookFormSchema = (t: TFunction) => z.object({
  endpointUrl: z
    .string()
    .min(1, t("webhooksSettingsPage.validation.endpointUrlRequired"))
    .url(t("webhooksSettingsPage.validation.invalidUrl"))
    .refine(
      (val) => {
        try {
          const url = new URL(val);
          return url.protocol === "https:";
        } catch {
          return false;
        }
      },
      t("webhooksSettingsPage.validation.httpsRequired"),
    ),
  selectedEvents: z
    .array(z.enum(WEBHOOK_EVENT_VALUES))
    .min(1, t("webhooksSettingsPage.validation.eventRequired")),
});

type WebhookFormData = z.infer<ReturnType<typeof createWebhookFormSchema>>;

interface CreateWebhookSubscriptionDialogProps {
  onCreated: () => void;
}

export function CreateWebhookSubscriptionDialog({
  onCreated,
}: CreateWebhookSubscriptionDialogProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const { register, handleSubmit, formState, setValue, watch, reset }
    = useFormWithSchema(createWebhookFormSchema(t), {
      defaultValues: {
        endpointUrl: "",
        selectedEvents: [],
      },
    });
  const selectedEvents = watch("selectedEvents");
  const [createWebhook, isCreating] = useMutation<CreateWebhookSubscriptionDialogMutation>(
    createWebhookSubscriptionMutation,
    {
      successMessage: t("webhooksSettingsPage.messages.created"),
      errorToast: t("webhooksSettingsPage.errors.create"),
    },
  );

  function handleToggleEvent(event: WebhookEventTypeValue) {
    const current = selectedEvents ?? [];
    const next = current.includes(event)
      ? current.filter(value => value !== event)
      : [...current, event];
    setValue("selectedEvents", next, { shouldValidate: formState.isSubmitted });
  }

  function onFormSubmit(data: WebhookFormData) {
    void createWebhook({
      variables: {
        input: {
          organizationId,
          endpointUrl: data.endpointUrl,
          selectedEvents: data.selectedEvents,
        },
      },
    }).then(
      () => {
        dialogRef.current?.close();
        reset({ endpointUrl: "", selectedEvents: [] });
        onCreated();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog
      ref={dialogRef}
      trigger={(
        <Button variant="solid" iconStart={<PlusIcon />}>
          {t("webhooksSettingsPage.actions.add")}
        </Button>
      )}
      title={t("webhooksSettingsPage.dialogs.addTitle")}
      className="max-w-lg"
    >
      <form onSubmit={e => void handleSubmit(onFormSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <Field
              label={t("webhooksSettingsPage.fields.endpointUrl")}
              error={formState.errors.endpointUrl?.message}
              required
            >
              <Input
                {...register("endpointUrl")}
                type="url"
                placeholder={t("webhooksSettingsPage.placeholders.endpointUrl")}
              />
            </Field>
            <div>
              <Label>{t("webhooksSettingsPage.fields.events")}</Label>
              <p className="text-sm text-txt-tertiary mb-2">
                {t("webhooksSettingsPage.eventsHelp")}
              </p>
              <div className="space-y-2">
                {WEBHOOK_EVENT_TYPES.map(event => (
                  <label
                    key={event.value}
                    className="flex items-center gap-2 cursor-pointer"
                  >
                    <Checkbox
                      checked={selectedEvents?.includes(event.value) ?? false}
                      onChange={() => handleToggleEvent(event.value)}
                    />
                    <span className="text-sm font-mono">{event.label}</span>
                  </label>
                ))}
              </div>
              {formState.errors.selectedEvents?.message && (
                <p className="text-xs text-red-600 mt-1">
                  {formState.errors.selectedEvents.message}
                </p>
              )}
            </div>
          </div>
        </DialogContent>
        <DialogFooter>
          <LegacyButton
            type="submit"
            disabled={isCreating}
          >
            {isCreating
              ? <Spinner size={16} />
              : t("webhooksSettingsPage.actions.create")}
          </LegacyButton>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
