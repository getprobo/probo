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
import { Button, Card, Field, IconPlusLarge, TabItem, Tabs, useDialogRef } from "@probo/ui";
import type { FocusEvent } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CompliancePortalUpdatesPage_updateMailingListMutation } from "#/__generated__/core/CompliancePortalUpdatesPage_updateMailingListMutation.graphql";
import type { CompliancePortalUpdatesPageQuery } from "#/__generated__/core/CompliancePortalUpdatesPageQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { CompliancePortalPageHeader } from "../_components/CompliancePortalPageHeader";

import { CompliancePortalMailingList } from "./_components/CompliancePortalMailingList";
import { CompliancePortalUpdatesList, type UpdateNode } from "./_components/CompliancePortalUpdatesList";
import { ComplianceUpdateFormDialog } from "./_components/ComplianceUpdateFormDialog";
import { NewCompliancePortalSubscriberDialog } from "./_components/NewCompliancePortalSubscriberDialog";

export const compliancePortalUpdatesPageQuery = graphql`
  query CompliancePortalUpdatesPageQuery($compliancePortalId: ID!) {
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
  mutation CompliancePortalUpdatesPage_updateMailingListMutation($input: UpdateMailingListInput!) {
    updateMailingList(input: $input) {
      mailingList {
        id
        replyTo
      }
    }
  }
`;

type Tab = "updates" | "subscribers";

export function CompliancePortalUpdatesPage(props: {
  queryRef: PreloadedQuery<CompliancePortalUpdatesPageQuery>;
}) {
  const { queryRef } = props;
  const { t } = useTranslation("organizations/compliance-portals");
  const title = t("updatesPage.title");
  usePageTitle(title);
  const subscriberDialogRef = useDialogRef();
  const newUpdateDialogRef = useDialogRef();
  const editUpdateDialogRef = useDialogRef();

  const [activeTab, setActiveTab] = useState<Tab>("updates");
  const [selectedUpdate, setSelectedUpdate] = useState<UpdateNode | null>(null);

  const { compliancePortal } = usePreloadedQuery<CompliancePortalUpdatesPageQuery>(
    compliancePortalUpdatesPageQuery,
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

  const [updateMailingList]
    = useMutation<CompliancePortalUpdatesPage_updateMailingListMutation>(
      updateMailingListMutation,
      {
        errorToast: t("updatesPage.errors.update"),
      },
    );

  function handleReplyToBlur(event: FocusEvent<HTMLInputElement>) {
    if (!mailingListId) {
      return;
    }
    const next = event.currentTarget.value.trim() || null;
    const current = mailingList?.replyTo || null;
    if (next === current) {
      return;
    }
    void updateMailingList({
      variables: {
        input: {
          id: mailingListId,
          replyTo: next,
        },
      },
    });
  }

  const handleEditUpdate = (update: UpdateNode) => {
    setSelectedUpdate({ ...update });
    editUpdateDialogRef.current?.open();
  };

  return (
    <div className="space-y-6">
      <CompliancePortalPageHeader
        title={title}
        description={t("updatesPage.description")}
      />
      {mailingListId && (
        <Card className="p-6 space-y-4">
          <div>
            <h3 className="text-base font-medium">{t("updatesPage.settings.title")}</h3>
            <p className="text-sm text-txt-tertiary">
              {t("updatesPage.settings.description")}
            </p>
          </div>
          <Field
            key={mailingListId}
            label={t("updatesPage.settings.replyTo")}
            type="email"
            placeholder={t("updatesPage.settings.replyToPlaceholder")}
            defaultValue={mailingList?.replyTo ?? ""}
            onBlur={handleReplyToBlur}
          />
        </Card>
      )}

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Tabs>
            <TabItem active={activeTab === "updates"} onClick={() => setActiveTab("updates")}>
              {t("updatesPage.tabs.updates")}
            </TabItem>
            <TabItem active={activeTab === "subscribers"} onClick={() => setActiveTab("subscribers")}>
              {t("updatesPage.tabs.subscribers")}
            </TabItem>
          </Tabs>

          {activeTab === "updates" && mailingListId && (
            <Button icon={IconPlusLarge} onClick={() => newUpdateDialogRef.current?.open()}>
              {t("updatesPage.actions.addUpdate")}
            </Button>
          )}
          {activeTab === "subscribers" && mailingListId && (
            <Button icon={IconPlusLarge} onClick={() => subscriberDialogRef.current?.open()}>
              {t("updatesPage.actions.addSubscriber")}
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
