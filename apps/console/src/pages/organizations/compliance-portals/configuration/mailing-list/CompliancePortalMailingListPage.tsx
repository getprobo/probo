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

import { Button, Card, Field, IconPlusLarge, Spinner, TabItem, Tabs, useDialogRef } from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CompliancePortalMailingListPage_updateMailingListMutation } from "#/__generated__/core/CompliancePortalMailingListPage_updateMailingListMutation.graphql";
import type { CompliancePortalMailingListPageQuery } from "#/__generated__/core/CompliancePortalMailingListPageQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { CompliancePortalMailingList } from "./_components/CompliancePortalMailingList";
import { CompliancePortalUpdatesList, type UpdateNode } from "./_components/CompliancePortalUpdatesList";
import { ComplianceUpdateFormDialog } from "./_components/ComplianceUpdateFormDialog";
import { NewCompliancePortalSubscriberDialog } from "./_components/NewCompliancePortalSubscriberDialog";

export const compliancePortalMailingListPageQuery = graphql`
  query CompliancePortalMailingListPageQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        id
        mailingList {
          id
          replyTo
          ...CompliancePortalUpdatesListFragment
        }
        ...CompliancePortalMailingListFragment
      }
    }
  }
`;

const updateMailingListMutation = graphql`
  mutation CompliancePortalMailingListPage_updateMailingListMutation($input: UpdateMailingListInput!) {
    updateMailingList(input: $input) {
      mailingList {
        id
        replyTo
      }
    }
  }
`;

type Tab = "updates" | "subscribers";

export default function CompliancePortalMailingListPage(props: {
  queryRef: PreloadedQuery<CompliancePortalMailingListPageQuery>;
}) {
  const { queryRef } = props;
  const { t } = useTranslation("organizations/compliance-portals");
  const subscriberDialogRef = useDialogRef();
  const newUpdateDialogRef = useDialogRef();
  const editUpdateDialogRef = useDialogRef();

  const [activeTab, setActiveTab] = useState<Tab>("updates");
  const [selectedUpdate, setSelectedUpdate] = useState<UpdateNode | null>(null);

  const { compliancePortal } = usePreloadedQuery<CompliancePortalMailingListPageQuery>(
    compliancePortalMailingListPageQuery,
    queryRef,
  );

  if (compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }

  const mailingList = compliancePortal.mailingList;
  const mailingListId = mailingList?.id;

  const subscriberConnectionId = mailingListId
    ? ConnectionHandler.getConnectionID(mailingListId, "CompliancePortalMailingList_subscribers")
    : null;

  const updatesConnectionId = mailingListId
    ? ConnectionHandler.getConnectionID(mailingListId, "CompliancePortalUpdatesList_updates")
    : null;

  const [replyTo, setReplyTo] = useState(mailingList?.replyTo ?? "");
  // This page stays mounted across portal switches, so re-seed the field when it
  // starts editing a different mailing list — otherwise saving without touching
  // the input overwrites the newly selected portal's reply-to.
  const [editedMailingListId, setEditedMailingListId] = useState(mailingListId);
  if (editedMailingListId !== mailingListId) {
    setEditedMailingListId(mailingListId);
    setReplyTo(mailingList?.replyTo ?? "");
  }

  const [updateMailingList, isUpdating]
    = useMutation<CompliancePortalMailingListPage_updateMailingListMutation>(
      updateMailingListMutation,
      {
        successMessage: t("mailingListPage.messages.updated"),
        errorToast: t("mailingListPage.errors.update"),
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

  const handleEditUpdate = (update: UpdateNode) => {
    setSelectedUpdate({ ...update });
    editUpdateDialogRef.current?.open();
  };

  return (
    <div className="space-y-6">
      {mailingListId && (
        <Card className="p-6 space-y-4">
          <div>
            <h3 className="text-base font-medium">{t("mailingListPage.settings.title")}</h3>
            <p className="text-sm text-txt-tertiary">
              {t("mailingListPage.settings.description")}
            </p>
          </div>
          <div className="flex items-end gap-3">
            <div className="flex-1">
              <Field
                label={t("mailingListPage.settings.replyTo")}
                type="email"
                placeholder={t("mailingListPage.settings.replyToPlaceholder")}
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
              {t("mailingListPage.actions.save")}
            </Button>
          </div>
        </Card>
      )}

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Tabs>
            <TabItem active={activeTab === "updates"} onClick={() => setActiveTab("updates")}>
              {t("mailingListPage.tabs.updates")}
            </TabItem>
            <TabItem active={activeTab === "subscribers"} onClick={() => setActiveTab("subscribers")}>
              {t("mailingListPage.tabs.subscribers")}
            </TabItem>
          </Tabs>

          {activeTab === "updates" && mailingListId && (
            <Button icon={IconPlusLarge} onClick={() => newUpdateDialogRef.current?.open()}>
              {t("mailingListPage.actions.addUpdate")}
            </Button>
          )}
          {activeTab === "subscribers" && mailingListId && (
            <Button icon={IconPlusLarge} onClick={() => subscriberDialogRef.current?.open()}>
              {t("mailingListPage.actions.addSubscriber")}
            </Button>
          )}
        </div>

        {activeTab === "updates" && mailingList && (
          <CompliancePortalUpdatesList
            fragmentRef={mailingList}
            onEdit={handleEditUpdate}
          />
        )}

        {activeTab === "subscribers" && (
          <CompliancePortalMailingList fragmentRef={compliancePortal} />
        )}
      </div>

      {mailingListId && updatesConnectionId && (
        <ComplianceUpdateFormDialog
          ref={newUpdateDialogRef}
          mailingListId={mailingListId}
          connectionId={updatesConnectionId}
        />
      )}

      <ComplianceUpdateFormDialog
        ref={editUpdateDialogRef}
        update={selectedUpdate}
      />

      {mailingListId && subscriberConnectionId && (
        <NewCompliancePortalSubscriberDialog
          ref={subscriberDialogRef}
          mailingListId={mailingListId}
          connectionId={subscriberConnectionId}
        />
      )}
    </div>
  );
}
