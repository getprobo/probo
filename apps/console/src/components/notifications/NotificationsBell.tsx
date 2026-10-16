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

import { BellIcon, BellSlashIcon, CaretRightIcon } from "@phosphor-icons/react";
import { catchResultValue, formatError, type GraphQLError } from "@probo/helpers";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type MouseEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, graphql, useLazyLoadQuery, usePaginationFragment, useRelayEnvironment } from "react-relay";
import { Link, useNavigate } from "react-router";

import type { NotificationsBell_viewer$data, NotificationsBell_viewer$key } from "#/__generated__/core/NotificationsBell_viewer.graphql";
import type { NotificationsBellMarkReadMutation } from "#/__generated__/core/NotificationsBellMarkReadMutation.graphql";
import type { NotificationsBellPaginationQuery } from "#/__generated__/core/NotificationsBellPaginationQuery.graphql";
import type { NotificationsBellQuery } from "#/__generated__/core/NotificationsBellQuery.graphql";
import { employeePortalHref } from "#/lib/employeePortalHref";
import { useMutation } from "#/lib/relay/useMutation";
import { navRail } from "#/pages/iam/organizations/_components/shell/variants";

import { notificationsPanel } from "./variants";

// Lives outside pages/iam: that tree compiles against the connect schema, and
// the inbox is on the console API. The nav rail wraps this in CoreRelayProvider.
const notificationTypes = [
  "DOCUMENT_SIGNATURE_REQUESTED",
  "DOCUMENT_APPROVAL_REQUESTED",
  "TASK_ASSIGNED",
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
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        employeePortals(
          first: 1
          orderBy: { field: CREATED_AT, direction: ASC }
        ) @catch(to: RESULT) {
          edges {
            node {
              id
            }
          }
        }
      }
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
          readAt
          createdAt
        }
      }
    }
    unreadNotifications: notifications(
      organizationId: $organizationId
      filter: { types: $types, unread: true }
      first: 1
    ) {
      totalCount
    }
  }
`;

const markReadMutation = graphql`
  mutation NotificationsBellMarkReadMutation($input: MarkNotificationReadInput!) {
    markNotificationRead(input: $input) {
      notification {
        id
        readAt
      }
    }
  }
`;

type NotificationNode = NotificationsBell_viewer$data["notifications"]["edges"][number]["node"];

interface NotificationsBellProps {
  organizationId: string;
}

export function NotificationsBell({ organizationId }: NotificationsBellProps) {
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
  const employeePortalId = data.organization?.__typename === "Organization"
    ? catchResultValue(
      data.organization.employeePortals,
      "cannot list employee portals",
    )?.edges[0]?.node.id ?? null
    : null;
  const [open, setOpen] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const listRef = useRef<HTMLUListElement>(null);
  const [readIds, setReadIds] = useState<ReadonlySet<string>>(new Set());
  const pendingRead = useRef(new Map<string, Promise<void>>());
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const toast = useToast();
  const [markRead] = useMutation<NotificationsBellMarkReadMutation>(markReadMutation);
  const rail = navRail();
  const panel = notificationsPanel();

  useEffect(() => {
    let subscription: { unsubscribe: () => void } | null = null;

    function refresh() {
      subscription?.unsubscribe();
      subscription = fetchQuery(environment, notificationsQuery, variables).subscribe({});
    }

    window.addEventListener("focus", refresh);
    return () => {
      window.removeEventListener("focus", refresh);
      subscription?.unsubscribe();
    };
  }, [environment, variables]);

  const label = t("notifications.label");
  const clearedOnScreen = edges.filter(
    ({ node }) => node.readAt == null && readIds.has(node.id),
  ).length;
  const unreadCount = Math.max(0, pagination.data.unreadNotifications.totalCount - clearedOnScreen);

  function isUnread(notification: NotificationNode) {
    return notification.readAt == null && !readIds.has(notification.id);
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

  const reportError = useCallback((error: GraphQLError) => {
    toast.add({
      title: t("common.error"),
      description: formatError(t("mutation.errors.commit"), error),
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

  function refreshList() {
    fetchQuery(environment, notificationsQuery, variables).subscribe({});
  }

  // Closing the panel acknowledges every notification that was on screen.
  function markVisibleRead(): Promise<void> {
    const ids = edges
      .filter(({ node }) => node.readAt == null)
      .map(({ node }) => node.id);
    const fresh = ids.filter(id => !pendingRead.current.has(id));
    if (fresh.length > 0) {
      rememberRead(fresh);
    }

    let reported = false;
    for (const id of fresh) {
      const pending = markRead({
        variables: {
          input: { notificationId: id },
        },
      }, { errorToast: false }).then(() => undefined).catch((error: GraphQLError) => {
        forgetRead([id]);
        if (!reported) {
          reported = true;
          reportError(error);
        }
      }).finally(() => {
        pendingRead.current.delete(id);
      });
      pendingRead.current.set(id, pending);
    }

    if (pendingRead.current.size === 0) {
      return Promise.resolve();
    }

    return Promise.all([...pendingRead.current.values()]).then(() => {
      refreshList();
    });
  }

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (next) {
      refreshList();
      return;
    }
    void markVisibleRead();
  }

  function handleOpen(event: MouseEvent<HTMLAnchorElement>, notification: NotificationNode) {
    const href = notificationHref(organizationId, employeePortalId, notification);
    const openInPlace = event.button === 0
      && !event.metaKey
      && !event.ctrlKey
      && !event.shiftKey
      && !event.altKey;
    if (openInPlace) {
      event.preventDefault();
    }

    const pending = markVisibleRead();
    setOpen(false);
    if (openInPlace) {
      void pending.finally(() => {
        followNotification(href);
      });
    }
  }

  function followNotification(
    href: { external: true; href: string } | { external: false; to: string },
  ) {
    if (href.external) {
      window.location.assign(href.href);
      return;
    }

    void navigate(href.to);
  }

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger
        ref={triggerRef}
        className={rail.item()}
        aria-label={unreadCount > 0
          ? `${label}, ${t("notifications.unread", { count: unreadCount })}`
          : label}
      >
        <span className={rail.icon()}>
          <span className={panel.iconWrap()}>
            <BellIcon size={20} weight="regular" aria-hidden />
            <UnreadCount count={unreadCount} />
          </span>
        </span>
        <Text size={2} className={rail.label()}>
          {label}
        </Text>
        <CaretRightIcon className={rail.caret()} />
      </PopoverTrigger>
      <PopoverPopup
        className={panel.popup()}
        side="right"
        align="end"
        sideOffset={12}
        alignOffset={() => railBottomOffset(triggerRef.current)}
        collisionAvoidance={{ side: "shift", align: "shift", fallbackAxisSide: "none" }}
      >
        <div className={panel.header()}>
          <Text size={2} weight="medium">
            {label}
          </Text>
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
                {edges.map(({ node }) => {
                  const href = notificationHref(organizationId, employeePortalId, node);
                  const unread = isUnread(node);
                  const body = (
                    <>
                      <UnreadDot unread={unread} />
                      <span className={panel.body()}>
                        <Text
                          size={2}
                          weight={unread ? "medium" : "regular"}
                          highContrast={unread}
                        >
                          {notificationMessage(node, t)}
                        </Text>
                        <Text size={1} color="faint">
                          {when.format(new Date(node.createdAt))}
                        </Text>
                      </span>
                    </>
                  );

                  return (
                    <li key={node.id}>
                      {href.external
                        ? (
                            <a href={href.href} className={panel.item()} onClick={event => handleOpen(event, node)}>
                              {body}
                            </a>
                          )
                        : (
                            <Link to={href.to} className={panel.item()} onClick={event => handleOpen(event, node)}>
                              {body}
                            </Link>
                          )}
                    </li>
                  );
                })}
                {pagination.hasNext && <li aria-hidden data-notification-page className={panel.sentinel()} />}
              </ul>
            )}
      </PopoverPopup>
    </Popover>
  );
}

// The bell sits above the account row. align="end" points the alignment axis
// upward, so the offset is negative to drop the panel onto the rail's bottom
// edge — the same edge the account menu uses.
function railBottomOffset(trigger: HTMLElement | null): number {
  const rail = trigger?.parentElement;
  if (trigger == null || rail == null) {
    return 0;
  }
  const paddingBottom = Number.parseFloat(getComputedStyle(rail).paddingBottom) || 0;
  const contentBottom = rail.getBoundingClientRect().bottom - paddingBottom;
  return trigger.getBoundingClientRect().bottom - contentBottom;
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

function notificationHref(
  organizationId: string,
  employeePortalId: string | null,
  notification: NotificationNode,
): { external: true; href: string } | { external: false; to: string } {
  if (notification.type === "TASK_ASSIGNED") {
    return {
      external: false,
      to: `/organizations/${organizationId}/governance/tasks/${notification.subjectId}`,
    };
  }

  const section = notification.type === "DOCUMENT_APPROVAL_REQUESTED"
    ? "approvals"
    : "signatures";

  if (employeePortalId != null) {
    return {
      external: true,
      href: employeePortalHref(employeePortalId, section, notification.subjectId),
    };
  }

  return {
    external: false,
    to: `/organizations/${organizationId}/governance/documents/${notification.subjectId}`,
  };
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
  }
}
