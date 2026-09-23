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
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { WebhookSubscriptionList_organization$key } from "#/__generated__/core/WebhookSubscriptionList_organization.graphql";
import type { WebhookSubscriptionListRefetchQuery } from "#/__generated__/core/WebhookSubscriptionListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { webhooksSettingsPage, webhookSubscriptionList } from "../variants";

import { WebhookSubscriptionListItem } from "./WebhookSubscriptionListItem";

export const WEBHOOK_SUBSCRIPTION_PAGE_SIZE = 15;

export const webhookSubscriptionListFragment = graphql`
  fragment WebhookSubscriptionList_organization on Organization
  @refetchable(queryName: "WebhookSubscriptionListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    canCreate: permission(action: "core:webhook-subscription:create")
    webhookSubscriptions(
      first: $first
      after: $after
      last: $last
      before: $before
    ) @connection(key: "WebhookSubscriptionList_webhookSubscriptions") {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...WebhookSubscriptionListItem_webhookSubscription
        }
      }
    }
  }
`;

interface WebhookSubscriptionListProps {
  organizationKey: WebhookSubscriptionList_organization$key;
}

export function WebhookSubscriptionList({ organizationKey }: WebhookSubscriptionListProps) {
  const { t } = useTranslation();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const [organization, refetch] = useRefetchableFragment<
    WebhookSubscriptionListRefetchQuery,
    WebhookSubscriptionList_organization$key
  >(webhookSubscriptionListFragment, organizationKey);

  const pageVariablesRef = useRef<CursorPaginationVariables>({
    first: WEBHOOK_SUBSCRIPTION_PAGE_SIZE,
    after: null,
    last: null,
    before: null,
  });

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    pageVariablesRef.current = variables;
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const { isPending: isPagePending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    organization.webhookSubscriptions.pageInfo,
    WEBHOOK_SUBSCRIPTION_PAGE_SIZE,
  );

  const edges = organization.webhookSubscriptions.edges;
  const pageInfo = organization.webhookSubscriptions.pageInfo;
  const isPending = isRefetchPending || isPagePending;
  const { header, intro } = webhooksSettingsPage();
  const { root, results, grid, empty, pager } = webhookSubscriptionList({ pending: isPending });

  function handleDeleted() {
    if (edges.length === 1 && pageInfo.hasPreviousPage) {
      goPrevious();
      return;
    }
    startRefetchTransition(() => {
      refetch(pageVariablesRef.current, { fetchPolicy: "network-only" });
    });
  }

  return (
    <>
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("nav.webhooks")}
          </Heading>
          <Text size={2} color="faint">
            {t("webhooksSettingsPage.description")}
          </Text>
        </div>
        {organization.canCreate && (
          <ButtonLink to="new" variant="solid" iconStart={<PlusIcon />}>
            {t("webhooksSettingsPage.actions.add")}
          </ButtonLink>
        )}
      </div>
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {t("webhooksSettingsPage.empty")}
                </Text>
              </div>
            </Card>
          )
        : (
            <div className={root()}>
              <div aria-busy={isPending} className={results()}>
                <div className={grid()}>
                  {edges.map(({ node }) => (
                    <WebhookSubscriptionListItem
                      key={node.id}
                      webhookSubscriptionKey={node}
                      onDeleted={handleDeleted}
                    />
                  ))}
                </div>
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
            </div>
          )}
    </>
  );
}
