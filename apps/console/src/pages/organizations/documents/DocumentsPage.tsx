import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconBell2,
  IconPlusLarge,
  PageHeader,
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
        documents(first: 50, orderBy: { field: TITLE, direction: ASC }) {
          edges {
            node {
              canSendSigningNotifications: permission(
                action: "core:document:send-signing-notifications"
              )
            }
          }
        }
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

  const canSendAnySignatureNotifications = organization.documents.edges.some(
    ({ node: { canSendSigningNotifications } }) => canSendSigningNotifications,
  );

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
          {organization.canCreateDocument && (
            <CreateDocumentDialog
              connection={documentListConnectionId}
              trigger={
                <Button icon={IconPlusLarge}>{__("New document")}</Button>
              }
            />
          )}
        </div>
      </PageHeader>
      <DocumentList
        fKey={organization}
        onConnectionIdChange={setDocumentListConnectionId}
      />
    </div>
  );
}
