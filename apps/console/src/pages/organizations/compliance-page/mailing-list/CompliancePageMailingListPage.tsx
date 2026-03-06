import { useTranslate } from "@probo/i18n";
import { Button, Card, Field, IconPlusLarge, Spinner, TabItem, Tabs, useDialogRef } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CompliancePageMailingListPage_updateMailingListMutation } from "#/__generated__/core/CompliancePageMailingListPage_updateMailingListMutation.graphql";
import type { CompliancePageMailingListPageQuery } from "#/__generated__/core/CompliancePageMailingListPageQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { CompliancePageMailingList } from "./_components/CompliancePageMailingList";
import { CompliancePageNewsList } from "./_components/CompliancePageNewsList";
import { EditComplianceNewsDialog } from "./_components/EditComplianceNewsDialog";
import { NewComplianceNewsDialog } from "./_components/NewComplianceNewsDialog";
import { NewCompliancePageSubscriberDialog } from "./_components/NewCompliancePageSubscriberDialog";

export const compliancePageMailingListPageQuery = graphql`
  query CompliancePageMailingListPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: trustCenter @required(action: THROW) {
          id
          mailingList {
            id
            replyTo
          }
          ...CompliancePageMailingListFragment
          ...CompliancePageNewsListFragment
        }
      }
    }
  }
`;

const updateMailingListMutation = graphql`
  mutation CompliancePageMailingListPage_updateMailingListMutation($input: UpdateMailingListInput!) {
    updateMailingList(input: $input) {
      mailingList {
        id
        replyTo
      }
    }
  }
`;

type NewsNode = {
  id: string;
  title: string;
  body: string;
  status: "DRAFT" | "SENT";
  createdAt: string;
  updatedAt: string;
};

type Tab = "updates" | "subscribers";

export function CompliancePageMailingListPage(props: {
  queryRef: PreloadedQuery<CompliancePageMailingListPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();
  const subscriberDialogRef = useDialogRef();
  const newNewsDialogRef = useDialogRef();
  const editNewsDialogRef = useDialogRef();

  const [activeTab, setActiveTab] = useState<Tab>("updates");
  const [selectedNews, setSelectedNews] = useState<NewsNode | null>(null);

  const { organization } = usePreloadedQuery<CompliancePageMailingListPageQuery>(
    compliancePageMailingListPageQuery,
    queryRef,
  );

  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const { compliancePage } = organization;
  const mailingList = compliancePage.mailingList;
  const mailingListId = mailingList?.id;
  const trustCenterId = compliancePage.id; // used for the news connection ID

  const subscriberConnectionId = mailingListId
    ? ConnectionHandler.getConnectionID(mailingListId, "CompliancePageMailingList_subscribers")
    : null;

  const newsConnectionId = ConnectionHandler.getConnectionID(
    trustCenterId,
    "CompliancePageNewsList_mailingListUpdates",
  );

  const [replyTo, setReplyTo] = useState(mailingList?.replyTo ?? "");

  const [updateMailingList, isUpdating] = useMutationWithToasts<CompliancePageMailingListPage_updateMailingListMutation>(
    updateMailingListMutation,
    {
      successMessage: __("Mailing list updated successfully"),
      errorMessage: __("Failed to update mailing list"),
    },
  );

  const handleSaveReplyTo = () => {
    if (!mailingListId) return;
    void updateMailingList({
      variables: {
        input: {
          id: mailingListId,
          replyTo: replyTo.trim() || null,
        },
      },
    });
  };

  const handleEditNews = (news: NewsNode) => {
    setSelectedNews({ ...news });
    editNewsDialogRef.current?.open();
  };

  return (
    <div className="space-y-6">
      {mailingListId && (
        <Card className="p-6 space-y-4">
          <div>
            <h3 className="text-base font-medium">{__("Settings")}</h3>
            <p className="text-sm text-txt-tertiary">
              {__("Configure how your mailing list behaves")}
            </p>
          </div>
          <div className="flex items-end gap-3">
            <div className="flex-1">
              <Field
                label={__("Reply-to email")}
                type="email"
                placeholder={__("security@example.com")}
                value={replyTo}
                onChange={e => setReplyTo(e.target.value)}
              />
            </div>
            <Button
              onClick={handleSaveReplyTo}
              disabled={isUpdating}
              className="shrink-0"
            >
              {isUpdating && <Spinner />}
              {__("Save")}
            </Button>
          </div>
        </Card>
      )}

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Tabs>
            <TabItem active={activeTab === "updates"} onClick={() => setActiveTab("updates")}>
              {__("News")}
            </TabItem>
            <TabItem active={activeTab === "subscribers"} onClick={() => setActiveTab("subscribers")}>
              {__("Subscribers")}
            </TabItem>
          </Tabs>

          {activeTab === "updates" && (
            <Button icon={IconPlusLarge} onClick={() => newNewsDialogRef.current?.open()}>
              {__("Add News")}
            </Button>
          )}
          {activeTab === "subscribers" && mailingListId && (
            <Button icon={IconPlusLarge} onClick={() => subscriberDialogRef.current?.open()}>
              {__("Add Subscriber")}
            </Button>
          )}
        </div>

        {activeTab === "updates" && (
          <CompliancePageNewsList
            fragmentRef={compliancePage}
            onEdit={handleEditNews}
          />
        )}

        {activeTab === "subscribers" && (
          <CompliancePageMailingList fragmentRef={compliancePage} />
        )}
      </div>

      {mailingListId && (
        <NewComplianceNewsDialog
          ref={newNewsDialogRef}
          mailingListId={mailingListId}
          connectionId={newsConnectionId}
        />
      )}

      <EditComplianceNewsDialog
        ref={editNewsDialogRef}
        news={selectedNews}
      />

      {mailingListId && subscriberConnectionId && (
        <NewCompliancePageSubscriberDialog
          ref={subscriberDialogRef}
          mailingListId={mailingListId}
          connectionId={subscriberConnectionId}
        />
      )}
    </div>
  );
}
