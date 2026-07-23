// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import {
  Button,
  IconPlusLarge,
  PageHeader,
  TabItem,
  Tabs,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { DocumentsPageQuery } from "#/__generated__/core/DocumentsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateDocumentDialog } from "./_components/CreateDocumentDialog";
import { DocumentList } from "./_components/DocumentList";

export const documentsPageQuery = graphql`
  query DocumentsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateDocument: permission(action: "core:document:create")
        ...DocumentListFragment @arguments(first: 50, order: { field: TITLE, direction: ASC })
      }
    }
  }
`;

export default function DocumentsPage(props: {
  queryRef: PreloadedQuery<DocumentsPageQuery>;
}) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation();

  const { organization } = usePreloadedQuery<DocumentsPageQuery>(
    documentsPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  usePageTitle(t("documentsPage.pageTitle"));

  const [tab, setTab] = useState<"ACTIVE" | "ARCHIVED">("ACTIVE");
  const [documentListConnectionId, setDocumentListConnectionId] = useState(
    ConnectionHandler.getConnectionID(
      organizationId,
      "DocumentsListQuery_documents",
      { orderBy: { direction: "ASC", field: "TITLE" } },
    ),
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("documentsPage.title")}
        description={t("documentsPage.description")}
      >
        <div className="flex gap-2">
          {organization.canCreateDocument && tab === "ACTIVE" && (
            <CreateDocumentDialog
              connection={documentListConnectionId}
              trigger={(
                <Button icon={IconPlusLarge}>
                  {t("documentsPage.actions.new")}
                </Button>
              )}
            />
          )}
        </div>
      </PageHeader>
      <Tabs>
        <TabItem active={tab === "ACTIVE"} onClick={() => setTab("ACTIVE")}>
          {t("documentsPage.tabs.active")}
        </TabItem>
        <TabItem active={tab === "ARCHIVED"} onClick={() => setTab("ARCHIVED")}>
          {t("documentsPage.tabs.archived")}
        </TabItem>
      </Tabs>
      <DocumentList
        fKey={organization}
        onConnectionIdChange={setDocumentListConnectionId}
        tab={tab}
      />
    </div>
  );
}
