import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconBell2,
  IconPlusLarge,
  PageHeader,
  TabItem,
  Tabs,
  useConfirm,
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
  const confirm = useConfirm();

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

  const handleResendSigningNotifications = () => {
    confirm(
      async () => {
        await sendSigningNotifications({
          variables: {
            input: { organizationId },
          },
        });
      },
      {
        title: __("Resend signing notifications"),
        message: __(
          "Signing notifications are automatically sent when a signature is requested. Are you sure you want to resend notifications to all pending signatories?",
        ),
        variant: "primary",
        label: __("Resend"),
      },
    );
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
              onClick={handleResendSigningNotifications}
            >
              {__("Resend signing notifications")}
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
