// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconBell2,
  IconPlusLarge,
  PageHeader,
  TabItem,
  Tabs,
} from "@probo/ui";
import { useState } from "react";
import {
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { DocumentsPageQuery } from "#/__generated__/core/DocumentsPageQuery.graphql";
import {
  useSendSigningNotificationsMutation,
} from "#/hooks/graph/DocumentGraph";
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
  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<DocumentsPageQuery>(
    documentsPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const [sendSigningNotifications] = useSendSigningNotificationsMutation();

  usePageTitle(__("Documents"));

  const [canSendAnySignatureNotifications, setCanSendAnySignatureNotifications] = useState(false);
  const [tab, setTab] = useState<"ACTIVE" | "ARCHIVED">("ACTIVE");
  const [documentListConnectionId, setDocumentListConnectionId] = useState(
    ConnectionHandler.getConnectionID(
      organizationId,
      "DocumentsListQuery_documents",
      { orderBy: { direction: "ASC", field: "TITLE" } },
    ),
  );

  const handleSendSigningNotifications = async () => {
    await sendSigningNotifications({
      variables: {
        input: { organizationId },
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Documents")}
        description={__("Manage your organization's documents")}
      >
        <div className="flex gap-2">
          {canSendAnySignatureNotifications && (
            <Button
              icon={IconBell2}
              variant="secondary"
              onClick={() => void handleSendSigningNotifications()}
            >
              {__("Send signing notifications")}
            </Button>
          )}
          {organization.canCreateDocument && tab === "ACTIVE" && (
            <CreateDocumentDialog
              connection={documentListConnectionId}
              trigger={
                <Button icon={IconPlusLarge}>{__("New document")}</Button>
              }
            />
          )}
        </div>
      </PageHeader>
      <Tabs>
        <TabItem active={tab === "ACTIVE"} onClick={() => setTab("ACTIVE")}>
          {__("Active")}
        </TabItem>
        <TabItem active={tab === "ARCHIVED"} onClick={() => setTab("ARCHIVED")}>
          {__("Archived")}
        </TabItem>
      </Tabs>
      <DocumentList
        fKey={organization}
        onConnectionIdChange={setDocumentListConnectionId}
        onCanSendNotificationsChange={setCanSendAnySignatureNotifications}
        tab={tab}
      />
    </div>
  );
}
