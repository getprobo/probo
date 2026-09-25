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

import { usePageTitle } from "@probo/hooks";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { WebhooksSettingsPageQuery } from "#/__generated__/core/WebhooksSettingsPageQuery.graphql";

import { WebhookSubscriptionList } from "./_components/WebhookSubscriptionList";
import { webhooksSettingsPage } from "./variants";

export const webhooksSettingsPageQuery = graphql`
  query WebhooksSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...WebhookSubscriptionList_organization
      }
    }
  }
`;

interface WebhooksSettingsPageProps {
  queryRef: PreloadedQuery<WebhooksSettingsPageQuery>;
}

export function WebhooksSettingsPage({ queryRef }: WebhooksSettingsPageProps) {
  const { t } = useTranslation();
  const { root } = webhooksSettingsPage();
  const title = t("nav.webhooks");
  usePageTitle(title);

  const { organization } = usePreloadedQuery<WebhooksSettingsPageQuery>(
    webhooksSettingsPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  return (
    <div className={root()}>
      <WebhookSubscriptionList organizationKey={organization} />
    </div>
  );
}
