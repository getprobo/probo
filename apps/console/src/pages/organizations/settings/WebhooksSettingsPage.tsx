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
import {
  Badge,
  Button,
  Card,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconPencil,
  IconPlusLarge,
  IconSquareBehindSquare2,
  IconTrashCan,
  Input,
  Label,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import type { TFunction } from "i18next";
import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery, useRelayEnvironment } from "react-relay";
import { ConnectionHandler, fetchQuery, graphql } from "relay-runtime";
import { z } from "zod";

import type { WebhooksSettingsPage_createMutation } from "#/__generated__/core/WebhooksSettingsPage_createMutation.graphql";
import type { WebhooksSettingsPage_deleteMutation } from "#/__generated__/core/WebhooksSettingsPage_deleteMutation.graphql";
import type { WebhooksSettingsPage_eventsQuery } from "#/__generated__/core/WebhooksSettingsPage_eventsQuery.graphql";
import type { WebhooksSettingsPage_signingSecretQuery } from "#/__generated__/core/WebhooksSettingsPage_signingSecretQuery.graphql";
import type { WebhooksSettingsPage_updateMutation } from "#/__generated__/core/WebhooksSettingsPage_updateMutation.graphql";
import type { WebhooksSettingsPageQuery } from "#/__generated__/core/WebhooksSettingsPageQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

export const webhooksSettingsPageQuery = graphql`
  query WebhooksSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        webhookSubscriptions(first: 50)
          @connection(key: "WebhooksSettingsPage_webhookSubscriptions") {
          edges {
            node {
              id
              endpointUrl
              selectedEvents
              events(first: 0) {
                totalCount
              }
            }
          }
        }
      }
    }
  }
`;

const createWebhookSubscriptionMutation = graphql`
  mutation WebhooksSettingsPage_createMutation(
    $input: CreateWebhookSubscriptionInput!
    $connections: [ID!]!
  ) {
    createWebhookSubscription(input: $input) {
      webhookSubscriptionEdge @prependEdge(connections: $connections) {
        node {
          id
          endpointUrl
          selectedEvents
          events(first: 0) {
            totalCount
          }
        }
      }
    }
  }
`;

const updateWebhookSubscriptionMutation = graphql`
  mutation WebhooksSettingsPage_updateMutation(
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

const signingSecretQuery = graphql`
  query WebhooksSettingsPage_signingSecretQuery($webhookSubscriptionId: ID!) {
    node(id: $webhookSubscriptionId) {
      ... on WebhookSubscription {
        signingSecret
      }
    }
  }
`;

const webhookEventsQuery = graphql`
  query WebhooksSettingsPage_eventsQuery(
    $webhookSubscriptionId: ID!
    $first: Int
    $after: CursorKey
  ) {
    node(id: $webhookSubscriptionId) {
      ... on WebhookSubscription {
        events(first: $first, after: $after) {
          totalCount
          pageInfo {
            hasNextPage
            endCursor
          }
          edges {
            node {
              id
              status
              createdAt
              response
            }
          }
        }
      }
    }
  }
`;

const deleteWebhookSubscriptionMutation = graphql`
  mutation WebhooksSettingsPage_deleteMutation(
    $input: DeleteWebhookSubscriptionInput!
    $connections: [ID!]!
  ) {
    deleteWebhookSubscription(input: $input) {
      deletedWebhookSubscriptionId @deleteEdge(connections: $connections)
    }
  }
`;

const EVENT_TYPES = [
  { value: "THIRD_PARTY_CREATED", label: "third-party:created" },
  { value: "THIRD_PARTY_UPDATED", label: "third-party:updated" },
  { value: "THIRD_PARTY_DELETED", label: "third-party:deleted" },
  { value: "USER_CREATED", label: "user:created" },
  { value: "USER_UPDATED", label: "user:updated" },
  { value: "USER_DELETED", label: "user:deleted" },
  { value: "OBLIGATION_CREATED", label: "obligation:created" },
  { value: "OBLIGATION_UPDATED", label: "obligation:updated" },
  { value: "OBLIGATION_DELETED", label: "obligation:deleted" },
  { value: "RIGHT_REQUEST_CREATED", label: "right-request:created" },
  { value: "RIGHT_REQUEST_UPDATED", label: "right-request:updated" },
  { value: "RIGHT_REQUEST_DELETED", label: "right-request:deleted" },
  { value: "DOCUMENT_CREATED", label: "document:created" },
  { value: "DOCUMENT_UPDATED", label: "document:updated" },
  { value: "DOCUMENT_ARCHIVED", label: "document:archived" },
  { value: "DOCUMENT_UNARCHIVED", label: "document:unarchived" },
  { value: "DOCUMENT_DELETED", label: "document:deleted" },
  { value: "DOCUMENT_VERSION_CREATED", label: "document-version:created" },
  { value: "DOCUMENT_VERSION_UPDATED", label: "document-version:updated" },
  { value: "DOCUMENT_VERSION_PUBLISHED", label: "document-version:published" },
  { value: "DOCUMENT_VERSION_REJECTED", label: "document-version:rejected" },
  { value: "DOCUMENT_VERSION_DELETED", label: "document-version:deleted" },
  { value: "DOCUMENT_VERSION_SIGNATURE_REQUESTED", label: "document-version-signature:requested" },
  { value: "DOCUMENT_VERSION_SIGNATURE_SIGNED", label: "document-version-signature:signed" },
  { value: "DOCUMENT_VERSION_SIGNATURE_CANCELLED", label: "document-version-signature:cancelled" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_REQUESTED", label: "document-version-approval-quorum:requested" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_UPDATED", label: "document-version-approval-quorum:updated" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_APPROVED", label: "document-version-approval-quorum:approved" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_REJECTED", label: "document-version-approval-quorum:rejected" },
  { value: "DOCUMENT_VERSION_APPROVAL_QUORUM_VOIDED", label: "document-version-approval-quorum:voided" },
] as const;

type WebhookEventType = (typeof EVENT_TYPES)[number]["value"];

const WEBHOOK_EVENT_VALUES = EVENT_TYPES.map(e => e.value) as [
  WebhookEventType,
  ...WebhookEventType[],
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

function WebhookFormDialog({
  mode,
  initialValues,
  onSubmit,
  isSubmitting,
  trigger,
}: {
  mode: "create" | "edit";
  initialValues?: WebhookFormData;
  onSubmit: (values: WebhookFormData) => void;
  isSubmitting: boolean;
  trigger: React.ReactNode;
}) {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const { register, handleSubmit, formState, setValue, watch, reset }
    = useFormWithSchema(createWebhookFormSchema(t), {
      defaultValues: {
        endpointUrl: initialValues?.endpointUrl ?? "",
        selectedEvents: initialValues?.selectedEvents ?? [],
      },
    });

  const selectedEvents = watch("selectedEvents");

  const handleToggleEvent = (event: WebhookEventType) => {
    const current = selectedEvents ?? [];
    const next = current.includes(event)
      ? current.filter(e => e !== event)
      : [...current, event];
    setValue("selectedEvents", next, { shouldValidate: formState.isSubmitted });
  };

  const onFormSubmit = (data: WebhookFormData) => {
    onSubmit(data);
    dialogRef.current?.close();
    reset(data);
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={trigger}
      title={
        mode === "create"
          ? t("webhooksSettingsPage.dialogs.addTitle")
          : t("webhooksSettingsPage.dialogs.editTitle")
      }
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
                {EVENT_TYPES.map(event => (
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
          <Button
            type="submit"
            disabled={isSubmitting}
          >
            {isSubmitting
              ? <Spinner size={16} />
              : mode === "create"
                ? t("webhooksSettingsPage.actions.create")
                : t("webhooksSettingsPage.actions.save")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function EventStatusBadge({ status }: { status: string }) {
  const { t } = useTranslation();
  if (status === "SUCCEEDED") {
    return <Badge variant="success" size="sm">{t("webhooksSettingsPage.status.succeeded")}</Badge>;
  }
  if (status === "PENDING") {
    return <Badge variant="info" size="sm">{t("webhooksSettingsPage.status.pending")}</Badge>;
  }
  return <Badge variant="danger" size="sm">{t("webhooksSettingsPage.status.failed")}</Badge>;
}

function WebhookEventsDialog({
  webhookSubscriptionId,
  endpointUrl,
  onClose,
}: {
  webhookSubscriptionId: string;
  endpointUrl: string;
  onClose: () => void;
}) {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const environment = useRelayEnvironment();
  const dialogRef = useDialogRef();
  type EventNode = NonNullable<WebhooksSettingsPage_eventsQuery["response"]["node"]["events"]>["edges"][number]["node"];
  const [events, setEvents] = useState<EventNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [endCursor, setEndCursor] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);

  const PAGE_SIZE = 20;

  const loadEvents = useCallback(
    async (after?: string | null) => {
      setLoading(true);
      try {
        const data = await fetchQuery<WebhooksSettingsPage_eventsQuery>(
          environment,
          webhookEventsQuery,
          {
            webhookSubscriptionId,
            first: PAGE_SIZE,
            after: after ?? null,
          },
        ).toPromise();

        const connection = data?.node?.events;
        if (connection) {
          const newEvents = connection.edges.map(e => e.node);
          setEvents(prev => after ? [...prev, ...newEvents] : newEvents);
          setHasNextPage(connection.pageInfo.hasNextPage);
          setEndCursor(connection.pageInfo.endCursor ?? null);
          setTotalCount(connection.totalCount);
        }
      } catch {
        toast({
          title: t("webhooksSettingsPage.errorTitle"),
          description: t("webhooksSettingsPage.errors.loadEvents"),
          variant: "error",
        });
      } finally {
        setLoading(false);
      }
    },
    [environment, webhookSubscriptionId, toast, t],
  );

  useEffect(() => {
    dialogRef.current?.open();
    const id = requestAnimationFrame(() => void loadEvents());
    return () => cancelAnimationFrame(id);
  }, [loadEvents, dialogRef]);

  return (
    <Dialog
      ref={dialogRef}
      title={t("webhooksSettingsPage.dialogs.eventsTitle")}
      className="max-w-2xl"
      onClose={onClose}
    >
      <DialogContent padded>
        <p className="text-sm text-txt-secondary mb-4">
          {endpointUrl}
          {totalCount > 0 && (
            <span className="text-txt-tertiary ml-2">
              {t("webhooksSettingsPage.total", { count: totalCount })}
            </span>
          )}
        </p>
        {events.length === 0 && !loading
          ? (
              <p className="text-sm text-txt-tertiary text-center py-8">
                {t("webhooksSettingsPage.emptyEvents")}
              </p>
            )
          : (
              <div className="space-y-2">
                {events.map(event => (
                  <div
                    key={event.id}
                    className="border border-border-solid rounded-md p-3 space-y-1"
                  >
                    <div className="flex items-center justify-between">
                      <EventStatusBadge status={event.status} />
                      <span className="text-xs text-txt-tertiary">
                        {dateTimeFormat(i18n.language, event.createdAt)}
                      </span>
                    </div>
                    {event.response && (
                      <details className="text-xs">
                        <summary className="cursor-pointer text-txt-link hover:underline">
                          {t("webhooksSettingsPage.response")}
                        </summary>
                        <pre className="mt-1 bg-subtle p-2 rounded text-xs overflow-auto max-h-48 whitespace-pre-wrap break-all">
                          {(() => {
                            try {
                              return JSON.stringify(JSON.parse(event.response), null, 2);
                            } catch {
                              return event.response;
                            }
                          })()}
                        </pre>
                      </details>
                    )}
                  </div>
                ))}
              </div>
            )}
        {loading && (
          <div className="flex justify-center py-4">
            <Spinner size={20} />
          </div>
        )}
      </DialogContent>
      {hasNextPage && !loading && (
        <DialogFooter>
          <Button
            variant="secondary"
            onClick={() => void loadEvents(endCursor)}
          >
            {t("webhooksSettingsPage.actions.loadMore")}
          </Button>
        </DialogFooter>
      )}
    </Dialog>
  );
}

export function WebhooksSettingsPage(props: {
  queryRef: PreloadedQuery<WebhooksSettingsPageQuery>;
}) {
  const { queryRef } = props;
  const { t } = useTranslation();
  const { toast } = useToast();
  const environment = useRelayEnvironment();
  const deleteDialogRef = useDialogRef();
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [revealedSecrets, setRevealedSecrets] = useState<Record<string, string>>({});
  const [loadingSecrets, setLoadingSecrets] = useState<Set<string>>(new Set());
  const [viewingEventsId, setViewingEventsId] = useState<string | null>(null);

  const fetchSigningSecret = useCallback(
    async (webhookSubscriptionId: string): Promise<string | null> => {
      // Return cached secret if already fetched
      if (revealedSecrets[webhookSubscriptionId]) {
        return revealedSecrets[webhookSubscriptionId];
      }

      setLoadingSecrets(prev => new Set(prev).add(webhookSubscriptionId));

      try {
        const data = await fetchQuery<WebhooksSettingsPage_signingSecretQuery>(
          environment,
          signingSecretQuery,
          { webhookSubscriptionId },
        ).toPromise();

        const secret = data?.node?.signingSecret;
        if (secret) {
          setRevealedSecrets(prev => ({ ...prev, [webhookSubscriptionId]: secret }));
          return secret;
        }
        return null;
      } catch {
        toast({
          title: t("webhooksSettingsPage.errorTitle"),
          description: t("webhooksSettingsPage.errors.loadSigningSecret"),
          variant: "error",
        });
        return null;
      } finally {
        setLoadingSecrets((prev) => {
          const next = new Set(prev);
          next.delete(webhookSubscriptionId);
          return next;
        });
      }
    },
    [environment, revealedSecrets, toast, t],
  );

  const toggleRevealSecret = (id: string) => {
    if (revealedSecrets[id]) {
      setRevealedSecrets((prev) => {
        const next = { ...prev };
        delete next[id];
        return next;
      });
    } else {
      void fetchSigningSecret(id);
    }
  };

  const copyToClipboard = async (webhookSubscriptionId: string, label: string) => {
    const secret = await fetchSigningSecret(webhookSubscriptionId);
    if (secret) {
      void navigator.clipboard.writeText(secret);
      toast({
        title: t("webhooksSettingsPage.copiedToClipboard"),
        description: label,
        variant: "success",
      });
    }
  };

  const { organization } = usePreloadedQuery<WebhooksSettingsPageQuery>(
    webhooksSettingsPageQuery,
    queryRef,
  );
  if (organization.__typename === "%other") {
    throw new Error("Relay node is not an organization");
  }

  const [createWebhook, isCreating]
    = useMutationWithToasts<WebhooksSettingsPage_createMutation>(
      createWebhookSubscriptionMutation,
      {
        successMessage: t("webhooksSettingsPage.messages.created"),
        errorMessage: t("webhooksSettingsPage.errors.create"),
      },
    );

  const [updateWebhook, isUpdating]
    = useMutationWithToasts<WebhooksSettingsPage_updateMutation>(
      updateWebhookSubscriptionMutation,
      {
        successMessage: t("webhooksSettingsPage.messages.updated"),
        errorMessage: t("webhooksSettingsPage.errors.update"),
      },
    );

  const [deleteWebhook, isDeleting]
    = useMutationWithToasts<WebhooksSettingsPage_deleteMutation>(
      deleteWebhookSubscriptionMutation,
      {
        successMessage: t("webhooksSettingsPage.messages.deleted"),
        errorMessage: t("webhooksSettingsPage.errors.delete"),
      },
    );

  const webhooks = organization.webhookSubscriptions?.edges ?? [];
  const viewingEventsWebhook = viewingEventsId
    ? webhooks.find(e => e.node.id === viewingEventsId)?.node ?? null
    : null;

  const connectionId = ConnectionHandler.getConnectionID(
    organization.id,
    "WebhooksSettingsPage_webhookSubscriptions",
  );

  const handleCreate = (values: WebhookFormData) => {
    void createWebhook({
      variables: {
        input: {
          organizationId: organization.id,
          endpointUrl: values.endpointUrl,
          selectedEvents: values.selectedEvents,
        },
        connections: [connectionId],
      },
    });
  };

  const handleUpdate = (id: string, values: WebhookFormData) => {
    void updateWebhook({
      variables: {
        input: {
          id,
          endpointUrl: values.endpointUrl,
          selectedEvents: values.selectedEvents,
        },
      },
    });
  };

  const handleDelete = (id: string) => {
    void deleteWebhook({
      variables: {
        input: {
          webhookSubscriptionId: id,
        },
        connections: [connectionId],
      },
      onSuccess: () => {
        setDeletingId(null);
        deleteDialogRef.current?.close();
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-base font-medium">{t("webhooksSettingsPage.title")}</h2>
          <p className="text-sm text-txt-tertiary">
            {t("webhooksSettingsPage.description")}
          </p>
        </div>
        <WebhookFormDialog
          mode="create"
          onSubmit={handleCreate}
          isSubmitting={isCreating}
          trigger={(
            <Button icon={IconPlusLarge}>
              {t("webhooksSettingsPage.actions.add")}
            </Button>
          )}
        />
      </div>

      {webhooks.length === 0
        ? (
            <Card padded>
              <div className="text-center py-8">
                <p className="text-sm text-txt-tertiary">
                  {t("webhooksSettingsPage.empty")}
                </p>
              </div>
            </Card>
          )
        : (
            <div className="space-y-3">
              {webhooks.map(({ node: webhook }) => (
                <Card key={webhook.id} padded>
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0 space-y-2">
                      <div>
                        <Label>{t("webhooksSettingsPage.fields.endpointUrl")}</Label>
                        <p className="text-sm font-mono text-txt-secondary truncate">
                          {webhook.endpointUrl}
                        </p>
                      </div>
                      <div>
                        <Label>{t("webhooksSettingsPage.fields.signingSecret")}</Label>
                        <div className="flex items-center gap-2 mt-1">
                          <code className="flex-1 bg-subtle p-2 rounded text-sm font-mono break-all">
                            {revealedSecrets[webhook.id]
                              ? revealedSecrets[webhook.id]
                              : "••••••••••••••••••••••••••••••••"}
                          </code>
                          <Button
                            variant="secondary"
                            onClick={() => toggleRevealSecret(webhook.id)}
                            disabled={loadingSecrets.has(webhook.id)}
                          >
                            {loadingSecrets.has(webhook.id)
                              ? <Spinner size={16} />
                              : revealedSecrets[webhook.id]
                                ? t("webhooksSettingsPage.actions.hide")
                                : t("webhooksSettingsPage.actions.show")}
                          </Button>
                          <Button
                            variant="secondary"
                            onClick={() => void copyToClipboard(webhook.id, t("webhooksSettingsPage.fields.signingSecret"))}
                            disabled={loadingSecrets.has(webhook.id)}
                            icon={IconSquareBehindSquare2}
                            aria-label={t("webhooksSettingsPage.copySigningSecret")}
                          />
                        </div>
                      </div>
                      <div>
                        <Label>{t("webhooksSettingsPage.fields.events")}</Label>
                        <div className="flex flex-wrap gap-1.5 mt-1">
                          {webhook.selectedEvents.map((event) => {
                            const eventLabel
                              = EVENT_TYPES.find(e => e.value === event)?.label ?? event;
                            return (
                              <span
                                key={event}
                                className="inline-flex items-center rounded-md bg-surface-secondary px-2 py-0.5 text-xs font-mono text-txt-secondary border border-border-solid"
                              >
                                {eventLabel}
                              </span>
                            );
                          })}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      <Button
                        variant="secondary"
                        onClick={() => setViewingEventsId(webhook.id)}
                      >
                        {t("webhooksSettingsPage.eventsCount", { count: webhook.events.totalCount })}
                      </Button>
                      <WebhookFormDialog
                        mode="edit"
                        initialValues={{
                          endpointUrl: webhook.endpointUrl,
                          selectedEvents: webhook.selectedEvents as WebhookEventType[],
                        }}
                        onSubmit={values => handleUpdate(webhook.id, values)}
                        isSubmitting={isUpdating}
                        trigger={(
                          <Button
                            variant="secondary"
                            icon={IconPencil}
                            aria-label={t("webhooksSettingsPage.editWebhook")}
                          />
                        )}
                      />
                      <Button
                        variant="quaternary"
                        icon={IconTrashCan}
                        aria-label={t("webhooksSettingsPage.deleteWebhook")}
                        className="text-red-600 hover:text-red-700"
                        onClick={() => {
                          setDeletingId(webhook.id);
                          deleteDialogRef.current?.open();
                        }}
                      />
                    </div>
                  </div>
                </Card>
              ))}
            </div>
          )}

      <Dialog
        ref={deleteDialogRef}
        title={t("webhooksSettingsPage.dialogs.deleteTitle")}
        className="max-w-md"
      >
        <DialogContent padded>
          <p className="text-txt-secondary">
            {t("webhooksSettingsPage.deleteConfirmation")}
          </p>
          <p className="text-txt-secondary mt-2">
            {t("webhooksSettingsPage.cannotUndo")}
          </p>
        </DialogContent>
        <DialogFooter>
          <Button
            variant="danger"
            onClick={() => deletingId && handleDelete(deletingId)}
            disabled={isDeleting}
            icon={isDeleting ? undefined : IconTrashCan}
          >
            {isDeleting
              ? (
                  <>
                    <Spinner size={16} />
                    {" "}
                    {t("webhooksSettingsPage.actions.deleting")}
                  </>
                )
              : t("webhooksSettingsPage.actions.delete")}
          </Button>
        </DialogFooter>
      </Dialog>

      {viewingEventsWebhook && viewingEventsId && (
        <WebhookEventsDialog
          webhookSubscriptionId={viewingEventsId}
          endpointUrl={viewingEventsWebhook.endpointUrl}
          onClose={() => setViewingEventsId(null)}
        />
      )}
    </div>
  );
}
