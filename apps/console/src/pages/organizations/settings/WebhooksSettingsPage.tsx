import { useTranslate } from "@probo/i18n";
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
import { useCallback, useEffect, useState } from "react";
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
  { value: "MEETING_CREATED", label: "meeting:created" },
  { value: "MEETING_UPDATED", label: "meeting:updated" },
  { value: "MEETING_DELETED", label: "meeting:deleted" },
  { value: "VENDOR_CREATED", label: "vendor:created" },
  { value: "VENDOR_UPDATED", label: "vendor:updated" },
  { value: "VENDOR_DELETED", label: "vendor:deleted" },
] as const;

type WebhookEventType = (typeof EVENT_TYPES)[number]["value"];

const WEBHOOK_EVENT_VALUES = EVENT_TYPES.map(e => e.value) as [
  WebhookEventType,
  ...WebhookEventType[],
];

const webhookFormSchema = z.object({
  endpointUrl: z
    .string()
    .min(1, "Endpoint URL is required")
    .url("Please enter a valid URL")
    .refine(
      (val) => {
        try {
          const url = new URL(val);
          return url.protocol === "https:";
        } catch {
          return false;
        }
      },
      "URL must use https://",
    ),
  selectedEvents: z
    .array(z.enum(WEBHOOK_EVENT_VALUES))
    .min(1, "At least one event must be selected"),
});

type WebhookFormData = z.infer<typeof webhookFormSchema>;

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
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const { register, handleSubmit, formState, setValue, watch, reset }
    = useFormWithSchema(webhookFormSchema, {
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
          ? __("Add Webhook Subscription")
          : __("Edit Webhook Subscription")
      }
      className="max-w-lg"
    >
      <form onSubmit={e => void handleSubmit(onFormSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <Field
              label={__("Endpoint URL")}
              error={formState.errors.endpointUrl?.message}
              required
            >
              <Input
                {...register("endpointUrl")}
                type="url"
                placeholder={__("https://example.com/webhook")}
              />
            </Field>
            <div>
              <Label>{__("Events")}</Label>
              <p className="text-sm text-txt-tertiary mb-2">
                {__("Select the events that will trigger this webhook.")}
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
                ? __("Create")
                : __("Save")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function EventStatusBadge({ status }: { status: string }) {
  const { __ } = useTranslate();
  if (status === "SUCCEEDED") {
    return <Badge variant="success" size="sm">{__("Succeeded")}</Badge>;
  }
  if (status === "PENDING") {
    return <Badge variant="info" size="sm">{__("Pending")}</Badge>;
  }
  return <Badge variant="danger" size="sm">{__("Failed")}</Badge>;
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString();
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
  const { __ } = useTranslate();
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
          title: __("Error"),
          description: __("Failed to load webhook events."),
          variant: "error",
        });
      } finally {
        setLoading(false);
      }
    },
    [environment, webhookSubscriptionId, toast, __],
  );

  useEffect(() => {
    void loadEvents();
    dialogRef.current?.open();
  }, [loadEvents]);

  return (
    <Dialog
      ref={dialogRef}
      title={__("Webhook Events")}
      className="max-w-2xl"
      onClose={onClose}
    >
      <DialogContent padded>
        <p className="text-sm text-txt-secondary mb-4">
          {endpointUrl}
          {totalCount > 0 && (
            <span className="text-txt-tertiary ml-2">
              {`(${totalCount} ${__("total")})`}
            </span>
          )}
        </p>
        {events.length === 0 && !loading
          ? (
              <p className="text-sm text-txt-tertiary text-center py-8">
                {__("No webhook events recorded yet.")}
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
                        {formatDate(event.createdAt)}
                      </span>
                    </div>
                    {event.response && (
                      <details className="text-xs">
                        <summary className="cursor-pointer text-txt-link hover:underline">
                          {__("Response")}
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
            {__("Load more")}
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
  const { __ } = useTranslate();
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
          title: __("Error"),
          description: __("Failed to load signing secret."),
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
    [environment, revealedSecrets, toast, __],
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
        title: __("Copied to clipboard"),
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
        successMessage: __("Webhook created successfully"),
        errorMessage: __("Failed to create webhook"),
      },
    );

  const [updateWebhook, isUpdating]
    = useMutationWithToasts<WebhooksSettingsPage_updateMutation>(
      updateWebhookSubscriptionMutation,
      {
        successMessage: __("Webhook updated successfully"),
        errorMessage: __("Failed to update webhook"),
      },
    );

  const [deleteWebhook, isDeleting]
    = useMutationWithToasts<WebhooksSettingsPage_deleteMutation>(
      deleteWebhookSubscriptionMutation,
      {
        successMessage: __("Webhook deleted successfully"),
        errorMessage: __("Failed to delete webhook"),
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
          <h2 className="text-base font-medium">{__("Webhook Subscriptions")}</h2>
          <p className="text-sm text-txt-tertiary">
            {__(
              "Configure webhooks to receive notifications when events occur in your organization.",
            )}
          </p>
        </div>
        <WebhookFormDialog
          mode="create"
          onSubmit={handleCreate}
          isSubmitting={isCreating}
          trigger={(
            <Button icon={IconPlusLarge}>
              {__("Add Webhook Subscription")}
            </Button>
          )}
        />
      </div>

      {webhooks.length === 0
        ? (
            <Card padded>
              <div className="text-center py-8">
                <p className="text-sm text-txt-tertiary">
                  {__("No webhook subscriptions yet. Add one to get started.")}
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
                        <Label>{__("Endpoint URL")}</Label>
                        <p className="text-sm font-mono text-txt-secondary truncate">
                          {webhook.endpointUrl}
                        </p>
                      </div>
                      <div>
                        <Label>{__("Signing Secret")}</Label>
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
                                ? __("Hide")
                                : __("Show")}
                          </Button>
                          <Button
                            variant="secondary"
                            onClick={() => void copyToClipboard(webhook.id, __("Signing Secret"))}
                            disabled={loadingSecrets.has(webhook.id)}
                            icon={IconSquareBehindSquare2}
                            aria-label={__("Copy signing secret")}
                          />
                        </div>
                      </div>
                      <div>
                        <Label>{__("Events")}</Label>
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
                        {`${__("Events")} (${webhook.events.totalCount})`}
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
                            aria-label={__("Edit webhook")}
                          />
                        )}
                      />
                      <Button
                        variant="quaternary"
                        icon={IconTrashCan}
                        aria-label={__("Delete webhook")}
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
        title={__("Delete Webhook")}
        className="max-w-md"
      >
        <DialogContent padded>
          <p className="text-txt-secondary">
            {__(
              "Are you sure you want to delete this webhook subscription?",
            )}
          </p>
          <p className="text-txt-secondary mt-2">
            {__("This action cannot be undone.")}
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
                    {__("Deleting...")}
                  </>
                )
              : __("Delete")}
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
