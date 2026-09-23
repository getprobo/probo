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

import { CaretDownIcon, CopyIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Collapsible } from "@probo/ui/src/v2/Collapsible/Collapsible";
import { CollapsiblePanel } from "@probo/ui/src/v2/Collapsible/CollapsiblePanel";
import { CollapsibleTrigger } from "@probo/ui/src/v2/Collapsible/CollapsibleTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { WebhookSubscriptionEventList_webhookSubscription$key } from "#/__generated__/core/WebhookSubscriptionEventList_webhookSubscription.graphql";
import type { WebhookSubscriptionEventListRefetchQuery } from "#/__generated__/core/WebhookSubscriptionEventListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import {
  useWebhookEventListFilters,
  webhookEventListGraphqlFilter,
} from "../_lib/useWebhookEventListFilters";
import { webhookSubscriptionEventList } from "../variants";

import { WebhookSubscriptionEventListFilter } from "./WebhookSubscriptionEventListFilter";

export const WEBHOOK_EVENT_PAGE_SIZE = 20;

export const webhookSubscriptionEventListFragment = graphql`
  fragment WebhookSubscriptionEventList_webhookSubscription on WebhookSubscription
  @refetchable(queryName: "WebhookSubscriptionEventListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "WebhookEventFilter", defaultValue: null }
  ) {
    events(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: CREATED_AT, direction: DESC }
      filter: $filter
    ) {
      totalCount
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          status
          createdAt
          payload
          response
        }
      }
    }
  }
`;

function headerValue(
  headers: Record<string, unknown>,
  name: string,
): string | undefined {
  const match = Object.entries(headers).find(
    ([key]) => key.toLowerCase() === name.toLowerCase(),
  );
  if (match == null) {
    return undefined;
  }

  const value = match[1];
  if (typeof value === "string" && value !== "") {
    return value;
  }
  if (Array.isArray(value) && typeof value[0] === "string" && value[0] !== "") {
    return value[0];
  }

  return undefined;
}

function parseResponse(response: string): {
  formatted: string;
  statusCode: number | undefined;
  contentType: string | undefined;
} {
  try {
    const parsed: unknown = JSON.parse(response);
    if (parsed == null || typeof parsed !== "object") {
      return { formatted: JSON.stringify(parsed, null, 2), statusCode: undefined, contentType: undefined };
    }

    const statusCode = "status_code" in parsed && typeof parsed.status_code === "number"
      ? parsed.status_code
      : undefined;
    const headers = "headers" in parsed
      && parsed.headers != null
      && typeof parsed.headers === "object"
      && !Array.isArray(parsed.headers)
      ? parsed.headers as Record<string, unknown>
      : undefined;
    const contentType = headers != null ? headerValue(headers, "Content-Type") : undefined;

    return {
      formatted: JSON.stringify(parsed, null, 2),
      statusCode,
      contentType,
    };
  } catch {
    return { formatted: response, statusCode: undefined, contentType: undefined };
  }
}

function EventStatusBadge({ status }: { status: string }) {
  const { t } = useTranslation();
  if (status === "SUCCEEDED") {
    return (
      <Badge variant="soft" color="green" size={2}>
        {t("webhooksSettingsPage.status.succeeded")}
      </Badge>
    );
  }
  if (status === "PENDING") {
    return (
      <Badge variant="soft" color="sky" size={2}>
        {t("webhooksSettingsPage.status.pending")}
      </Badge>
    );
  }
  return (
    <Badge variant="soft" color="red" size={2}>
      {t("webhooksSettingsPage.status.failed")}
    </Badge>
  );
}

function formatJSON(value: string): string {
  try {
    return JSON.stringify(JSON.parse(value) as unknown, null, 2);
  } catch {
    return value;
  }
}

function DeliveryJsonBlock({
  copyError,
  copyLabel,
  label,
  value,
}: {
  copyError: string;
  copyLabel: string;
  label: string;
  value: string;
}) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const {
    block,
    blockHeading,
    responseWrap,
    response: responseClass,
    responseCopy,
  } = webhookSubscriptionEventList();

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(value);
      toast({
        title: t("webhooksSettingsPage.copiedToClipboard"),
        description: label,
        variant: "success",
      });
    } catch {
      toast({
        title: t("webhooksSettingsPage.errorTitle"),
        description: copyError,
        variant: "error",
      });
    }
  }

  return (
    <div className={block()}>
      <Text size={1} color="faint" className={blockHeading()}>
        {label}
      </Text>
      <div className={responseWrap()}>
        <IconButton
          variant="surface"
          color="neutral"
          size={1}
          className={responseCopy()}
          aria-label={copyLabel}
          onClick={() => {
            void handleCopy();
          }}
        >
          <CopyIcon />
        </IconButton>
        <pre className={responseClass()}>
          {value}
        </pre>
      </div>
    </div>
  );
}

function DeliveryRow({
  createdAt,
  payload,
  response,
  status,
}: {
  createdAt: string;
  payload: string | null | undefined;
  response: string | null | undefined;
  status: string;
}) {
  const { t, i18n } = useTranslation();
  const {
    row,
    trigger,
    lead,
    trail,
    contentType,
    caret,
    panel,
  } = webhookSubscriptionEventList();
  const formattedPayload = payload != null && payload !== "" ? formatJSON(payload) : null;
  const parsed = response != null && response !== "" ? parseResponse(response) : null;
  const expandable = formattedPayload != null || parsed != null;

  const header = (
    <>
      <div className={lead()}>
        <Text size={2} color="faint">
          {dateTimeFormat(i18n.language, createdAt)}
        </Text>
        <EventStatusBadge status={status} />
        {parsed?.statusCode != null && (
          <Code variant="ghost">{String(parsed.statusCode)}</Code>
        )}
      </div>
      <div className={trail()}>
        {parsed?.contentType != null && (
          <Code variant="ghost" className={contentType()}>
            {parsed.contentType}
          </Code>
        )}
        {expandable && <CaretDownIcon className={caret()} aria-hidden />}
      </div>
    </>
  );

  if (!expandable) {
    return <div className={trigger()}>{header}</div>;
  }

  return (
    <Collapsible className={row()}>
      <CollapsibleTrigger
        className={trigger()}
        aria-label={formattedPayload != null
          ? t("webhooksSettingsPage.payload")
          : t("webhooksSettingsPage.response")}
      >
        {header}
      </CollapsibleTrigger>
      <CollapsiblePanel>
        <div className={panel()}>
          {formattedPayload != null && (
            <DeliveryJsonBlock
              label={t("webhooksSettingsPage.payload")}
              copyLabel={t("webhooksSettingsPage.copyPayload")}
              copyError={t("webhooksSettingsPage.errors.copyPayload")}
              value={formattedPayload}
            />
          )}
          {parsed != null && (
            <DeliveryJsonBlock
              label={t("webhooksSettingsPage.response")}
              copyLabel={t("webhooksSettingsPage.copyResponse")}
              copyError={t("webhooksSettingsPage.errors.copyResponse")}
              value={parsed.formatted}
            />
          )}
        </div>
      </CollapsiblePanel>
    </Collapsible>
  );
}

interface WebhookSubscriptionEventListProps {
  webhookSubscriptionKey: WebhookSubscriptionEventList_webhookSubscription$key;
}

export function WebhookSubscriptionEventList({
  webhookSubscriptionKey,
}: WebhookSubscriptionEventListProps) {
  const { t } = useTranslation();
  const { status, hasActiveFilters } = useWebhookEventListFilters();
  const [isFilterPending, startTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [webhook, refetch] = useRefetchableFragment<
    WebhookSubscriptionEventListRefetchQuery,
    WebhookSubscriptionEventList_webhookSubscription$key
  >(webhookSubscriptionEventListFragment, webhookSubscriptionKey);

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const { isPending: isPagePending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    webhook.events.pageInfo,
    WEBHOOK_EVENT_PAGE_SIZE,
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    startTransition(() => {
      refetch(
        {
          first: WEBHOOK_EVENT_PAGE_SIZE,
          after: null,
          last: null,
          before: null,
          filter: webhookEventListGraphqlFilter(status),
        },
        { fetchPolicy: "network-only" },
      );
    });
  }, [status, refetch]);

  const isPending = isFilterPending || isPagePending;
  const edges = webhook.events.edges;
  const pageInfo = webhook.events.pageInfo;
  const {
    root,
    heading,
    results,
    empty,
    item,
    pager,
  } = webhookSubscriptionEventList({ pending: isPending });

  return (
    <div className={root()}>
      <div className={heading()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("webhooksSettingsPage.deliveriesCount", { count: webhook.events.totalCount })}
        </Heading>
        <WebhookSubscriptionEventListFilter />
      </div>
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {hasActiveFilters
                    ? t("webhooksSettingsPage.emptyDeliveriesFiltered")
                    : t("webhooksSettingsPage.emptyDeliveries")}
                </Text>
              </div>
            </Card>
          )
        : (
            <>
              <div aria-busy={isPending} className={results()}>
                <List>
                  {edges.map(({ node }) => (
                    <ListItem key={node.id} className={item()}>
                      <DeliveryRow
                        status={node.status}
                        createdAt={node.createdAt}
                        payload={node.payload}
                        response={node.response}
                      />
                    </ListItem>
                  ))}
                </List>
              </div>
              <div className={pager()}>
                <Pagination
                  hasPrevious={pageInfo.hasPreviousPage}
                  hasNext={pageInfo.hasNextPage}
                  previousLabel={t("webhooksSettingsPage.actions.previous")}
                  nextLabel={t("webhooksSettingsPage.actions.next")}
                  showLabels
                  variant="surface"
                  disabled={isPending}
                  onPrevious={goPrevious}
                  onNext={goNext}
                />
              </div>
            </>
          )}
    </div>
  );
}
