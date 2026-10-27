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

import { BellIcon, BellSlashIcon, ChecksIcon, GearIcon } from "@phosphor-icons/react";
import { formatError, type GraphQLError } from "@probo/helpers";
import { iconButton } from "@probo/ui/src/v2/IconButton/variants";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, graphql, useLazyLoadQuery, usePaginationFragment, useRelayEnvironment } from "react-relay";
import { Link } from "react-router";

import type { NotificationsBell_viewer$data, NotificationsBell_viewer$key } from "#/__generated__/core/NotificationsBell_viewer.graphql";
import type { NotificationsBellMarkAllReadMutation } from "#/__generated__/core/NotificationsBellMarkAllReadMutation.graphql";
import type { NotificationsBellPaginationQuery } from "#/__generated__/core/NotificationsBellPaginationQuery.graphql";
import type { NotificationsBellQuery } from "#/__generated__/core/NotificationsBellQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { notificationsRefreshEvent } from "./requestNotificationRefresh";
import { notificationsPanel } from "./variants";

// Lives outside pages/iam: that tree compiles against the connect schema, and
// the inbox is on the console API. The top bar wraps this in the core Relay
// environment.
const notificationTypes = [
  "DOCUMENT_SIGNATURE_REQUESTED",
  "DOCUMENT_APPROVAL_REQUESTED",
  "TASK_ASSIGNED",
  "TASK_MENTIONED",
] as const;

const PAGE_SIZE = 10;

const notificationsQuery = graphql`
  query NotificationsBellQuery(
    $organizationId: ID!
    $types: [NotificationType!]
    $count: Int
  ) {
    viewer {
      ...NotificationsBell_viewer @arguments(
        organizationId: $organizationId
        types: $types
        count: $count
      )
    }
  }
`;

const notificationsFragment = graphql`
  fragment NotificationsBell_viewer on Viewer
  @refetchable(queryName: "NotificationsBellPaginationQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    types: { type: "[NotificationType!]" }
    count: { type: "Int", defaultValue: 10 }
    cursor: { type: "CursorKey", defaultValue: null }
  ) {
    id
    notifications(
      organizationId: $organizationId
      filter: { types: $types }
      first: $count
      after: $cursor
    ) @connection(key: "NotificationsBell_notifications") {
      edges {
        node {
          id
          type
          subjectId
          title
          status
          readAt
          createdAt
        }
      }
    }
    unreadNotifications: notifications(
      organizationId: $organizationId
      filter: { types: $types, statuses: [PENDING] }
      first: 1
    ) {
      totalCount
    }
  }
`;

const markAllReadMutation = graphql`
  mutation NotificationsBellMarkAllReadMutation($input: MarkAllNotificationsReadInput!) {
    markAllNotificationsRead(input: $input) {
      organizationId
    }
  }
`;

type NotificationNode = NotificationsBell_viewer$data["notifications"]["edges"][number]["node"];

interface NotificationsBellProps {
  organizationId: string;
  employeePortalId: string;
}

export function NotificationsBell({
  organizationId,
  employeePortalId,
}: NotificationsBellProps) {
  const variables = useMemo(
    () => ({ organizationId, types: notificationTypes, count: PAGE_SIZE }),
    [organizationId],
  );
  const environment = useRelayEnvironment();
  const data = useLazyLoadQuery<NotificationsBellQuery>(
    notificationsQuery,
    variables,
    { fetchPolicy: "store-and-network" },
  );
  const pagination = usePaginationFragment<
    NotificationsBellPaginationQuery,
    NotificationsBell_viewer$key
  >(notificationsFragment, data.viewer);
  const edges = pagination.data.notifications.edges;
  const [open, setOpen] = useState(false);
  const listRef = useRef<HTMLUListElement>(null);
  const [readIds, setReadIds] = useState<ReadonlySet<string>>(new Set());
  const [clearedThrough, setClearedThrough] = useState<{
    organizationId: string;
    count: number;
  } | null>(null);
  const markingAll = useRef(false);
  const { t, i18n } = useTranslation();
  const toast = useToast();
  const [markAllRead] = useMutation<NotificationsBellMarkAllReadMutation>(markAllReadMutation);
  const panel = notificationsPanel();

  useEffect(() => {
    let subscription: { unsubscribe: () => void } | null = null;

    function refresh() {
      subscription?.unsubscribe();
      subscription = fetchQuery(environment, notificationsQuery, variables).subscribe({});
    }

    window.addEventListener("focus", refresh);
    window.addEventListener(notificationsRefreshEvent, refresh);
    return () => {
      window.removeEventListener("focus", refresh);
      window.removeEventListener(notificationsRefreshEvent, refresh);
      subscription?.unsubscribe();
    };
  }, [environment, variables]);

  const label = t("notifications.label");
  const serverUnread = pagination.data.unreadNotifications.totalCount;
  const markedAll = clearedThrough != null
    && clearedThrough.organizationId === organizationId
    && clearedThrough.count === serverUnread;
  const unreadCount = markedAll ? 0 : serverUnread;

  function isUnread(notification: NotificationNode) {
    return !markedAll
      && notification.status === "PENDING"
      && !readIds.has(notification.id);
  }

  function rememberRead(ids: string[]) {
    setReadIds((current) => {
      const next = new Set(current);
      for (const id of ids) {
        next.add(id);
      }
      return next;
    });
  }

  function forgetRead(ids: string[]) {
    setReadIds((current) => {
      const next = new Set(current);
      for (const id of ids) {
        next.delete(id);
      }
      return next;
    });
  }
  const when = new Intl.DateTimeFormat(i18n.language, {
    dateStyle: "medium",
    timeStyle: "short",
  });

  function refreshList() {
    fetchQuery(environment, notificationsQuery, variables).subscribe({});
  }

  const reportError = useCallback((error: GraphQLError) => {
    toast.add({
      title: t("common.error"),
      description: formatError(t("common.error"), error),
      type: "error",
    });
  }, [toast, t]);

  const pageErrorRef = useRef(false);
  const requestingRef = useRef(false);
  const { hasNext, isLoadingNext, loadNext } = pagination;

  useEffect(() => {
    if (!open) {
      pageErrorRef.current = false;
      requestingRef.current = false;
      return;
    }
    const root = listRef.current;
    const sentinel = root?.querySelector("[data-notification-page]");
    if (isLoadingNext || root == null || !(sentinel instanceof HTMLElement)) {
      return;
    }

    const observer = new IntersectionObserver((entries) => {
      if (!entries.some(entry => entry.isIntersecting)) {
        return;
      }
      if (!hasNext || requestingRef.current || pageErrorRef.current) {
        return;
      }
      requestingRef.current = true;
      loadNext(PAGE_SIZE, {
        onComplete: (error) => {
          requestingRef.current = false;
          if (error != null) {
            pageErrorRef.current = true;
            reportError(error);
          }
        },
      });
    }, { root });
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [open, hasNext, isLoadingNext, loadNext, edges.length, reportError]);

  function markEveryRead() {
    if (markingAll.current || unreadCount === 0) {
      return;
    }

    const ids = edges
      .filter(({ node }) => isUnread(node))
      .map(({ node }) => node.id);
    markingAll.current = true;
    setClearedThrough({ organizationId, count: serverUnread });
    rememberRead(ids);
    void markAllRead({
      variables: {
        input: { organizationId, types: [...notificationTypes] },
      },
    }, { errorToast: false }).then(() => {
      refreshList();
    }).catch((error: GraphQLError) => {
      setClearedThrough(null);
      forgetRead(ids);
      reportError(error);
    }).finally(() => {
      markingAll.current = false;
    });
  }

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (next) {
      refreshList();
    }
  }

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <span className={panel.triggerWrap()}>
        <PopoverTrigger
          className={iconButton({
            variant: "ghost",
            color: "neutral",
            className: panel.trigger(),
          })}
          aria-label={unreadCount > 0
            ? `${label}, ${t("notifications.unread", { count: unreadCount })}`
            : label}
        >
          <BellIcon />
          <UnreadCount count={unreadCount} />
        </PopoverTrigger>
      </span>
      <PopoverPopup
        className={panel.popup()}
        align="end"
        sideOffset={8}
        collisionPadding={12}
      >
        <div className={panel.header()}>
          <Text size={2} weight="medium">
            {label}
          </Text>
          <div className={panel.actions()}>
            <Tooltip>
              <TooltipTrigger
                render={(
                  <button
                    type="button"
                    aria-label={t("notifications.markAllRead")}
                    className={panel.action()}
                    onClick={markEveryRead}
                  >
                    <ChecksIcon size={16} aria-hidden />
                  </button>
                )}
              />
              <TooltipPopup side="bottom">
                {t("notifications.markAllRead")}
              </TooltipPopup>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger
                render={(
                  <Link
                    to={`/${employeePortalId}/notification-settings`}
                    aria-label={t("notifications.settings")}
                    className={panel.action()}
                  >
                    <GearIcon size={16} />
                  </Link>
                )}
              />
              <TooltipPopup side="bottom">
                {t("notifications.settings")}
              </TooltipPopup>
            </Tooltip>
          </div>
        </div>
        {edges.length === 0
          ? (
              <div className={panel.empty()}>
                <BellSlashIcon size={20} weight="regular" className={panel.emptyIcon()} aria-hidden />
                <Heading level={2} size={2} weight="medium">
                  {t("notifications.empty")}
                </Heading>
              </div>
            )
          : (
              <ul ref={listRef} className={panel.list()}>
                {edges.map(({ node }) => (
                  <li key={node.id}>
                    <Link
                      to={notificationPath(employeePortalId, node)}
                      onClick={() => setOpen(false)}
                      className={panel.item()}
                    >
                      <UnreadDot unread={isUnread(node)} />
                      <span className={panel.body()}>
                        <Text
                          size={2}
                          weight={isUnread(node) ? "medium" : "regular"}
                          highContrast={isUnread(node)}
                        >
                          {notificationMessage(node, t)}
                        </Text>
                        <Text size={1} color="faint">
                          {when.format(new Date(node.createdAt))}
                        </Text>
                      </span>
                    </Link>
                  </li>
                ))}
                {pagination.hasNext && <li aria-hidden data-notification-page className={panel.sentinel()} />}
              </ul>
            )}
      </PopoverPopup>
    </Popover>
  );
}

function UnreadCount({ count }: { count: number }) {
  if (count <= 0) {
    return null;
  }

  return (
    <span className={notificationsPanel().count()}>
      <Text size={1} weight="medium" color="current">
        {count > 99 ? "99+" : count}
      </Text>
    </span>
  );
}

function UnreadDot({ unread }: { unread: boolean }) {
  return (
    <span aria-hidden className={notificationsPanel({ unread }).dot()} />
  );
}

function notificationPath(employeePortalId: string, notification: NotificationNode): string {
  if (notification.type === "DOCUMENT_SIGNATURE_REQUESTED") {
    return `/${employeePortalId}/signatures/${notification.subjectId}`;
  }
  if (notification.type === "DOCUMENT_APPROVAL_REQUESTED") {
    return `/${employeePortalId}/approvals/${notification.subjectId}`;
  }
  if (notification.type === "TASK_ASSIGNED" || notification.type === "TASK_MENTIONED") {
    return `/${employeePortalId}/tasks/${notification.subjectId}`;
  }

  return `/${employeePortalId}`;
}

function notificationMessage(
  notification: NotificationNode,
  t: (key: string, options?: { title: string }) => string,
): string {
  switch (notification.type) {
    case "DOCUMENT_SIGNATURE_REQUESTED":
      return t("notifications.documentSignatureRequested", { title: notification.title });
    case "DOCUMENT_APPROVAL_REQUESTED":
      return t("notifications.documentApprovalRequested", { title: notification.title });
    case "TASK_ASSIGNED":
      return t("notifications.taskAssigned", { title: notification.title });
    case "TASK_MENTIONED":
      return t("notifications.taskMentioned", { title: notification.title });
  }
}
