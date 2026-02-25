import { useTranslate } from "@probo/i18n";
import { Button, Card, Field, IconPlusLarge, Spinner, useDialogRef } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CompliancePageMailingListPage_updateMailingListMutation } from "#/__generated__/core/CompliancePageMailingListPage_updateMailingListMutation.graphql";
import type { CompliancePageMailingListPageQuery } from "#/__generated__/core/CompliancePageMailingListPageQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { CompliancePageMailingList } from "./_components/CompliancePageMailingList";
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

export function CompliancePageMailingListPage(props: {
  queryRef: PreloadedQuery<CompliancePageMailingListPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();

  const { organization } = usePreloadedQuery<CompliancePageMailingListPageQuery>(
    compliancePageMailingListPageQuery,
    queryRef,
  );

  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const mailingList = organization.compliancePage.mailingList;
  const mailingListId = mailingList?.id;

  const connectionId = mailingListId
    ? ConnectionHandler.getConnectionID(mailingListId, "CompliancePageMailingList_subscribers")
    : null;

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
          <div>
            <h3 className="text-base font-medium">{__("Subscribers")}</h3>
            <p className="text-sm text-txt-tertiary">
              {__("People subscribed to receive security and compliance updates")}
            </p>
          </div>
          {mailingListId && (
            <Button icon={IconPlusLarge} onClick={() => dialogRef.current?.open()}>
              {__("Add Subscriber")}
            </Button>
          )}
        </div>

        <CompliancePageMailingList fragmentRef={organization.compliancePage} />
      </div>

      {mailingListId && connectionId && (
        <NewCompliancePageSubscriberDialog
          ref={dialogRef}
          mailingListId={mailingListId}
          connectionId={connectionId}
        />
      )}
    </div>
  );
}
