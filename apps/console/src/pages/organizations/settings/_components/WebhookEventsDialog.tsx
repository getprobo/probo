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
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useRelayEnvironment } from "react-relay";
import { fetchQuery, graphql } from "relay-runtime";

import type { WebhookEventsDialogQuery } from "#/__generated__/core/WebhookEventsDialogQuery.graphql";

const webhookEventsQuery = graphql`
  query WebhookEventsDialogQuery(
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

export function WebhookEventsDialog({
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
  type EventNode = NonNullable<WebhookEventsDialogQuery["response"]["node"]["events"]>["edges"][number]["node"];
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
        const data = await fetchQuery<WebhookEventsDialogQuery>(
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
