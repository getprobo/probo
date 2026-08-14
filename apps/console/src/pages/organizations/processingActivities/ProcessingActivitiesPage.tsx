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
  Card,
  IconPageTextLine,
  IconPlusLarge,
  IconUpload,
  PageHeader,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";

import type { ProcessingActivitiesPageFragment$key } from "#/__generated__/core/ProcessingActivitiesPageFragment.graphql";
import type { ProcessingActivityGraphListQuery } from "#/__generated__/core/ProcessingActivityGraphListQuery.graphql";
import {
  ProcessingActivitiesConnectionKey,
  processingActivitiesQuery,
} from "#/hooks/graph/ProcessingActivityGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { ProcessingActivityListItem } from "./_components/ProcessingActivityListItem";
import { CreateProcessingActivityDialog } from "./dialogs/CreateProcessingActivityDialog";
import { PublishProcessingActivityListDialog } from "./dialogs/PublishProcessingActivityListDialog";

interface ProcessingActivitiesPageProps {
  queryRef: PreloadedQuery<ProcessingActivityGraphListQuery>;
}

const processingActivitiesPageFragment = graphql`
  fragment ProcessingActivitiesPageFragment on Organization
  @refetchable(queryName: "ProcessingActivitiesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
  ) {
    id
    processingActivities(first: $first, after: $after)
      @connection(key: "ProcessingActivitiesPage_processingActivities") {
      edges {
        node {
          id
          canUpdate: permission(action: "core:processing-activity:update")
          canDelete: permission(action: "core:processing-activity:delete")
          ...ProcessingActivityListItem_processingActivity
        }
      }
    }
  }
`;

export default function ProcessingActivitiesPage({
  queryRef,
}: ProcessingActivitiesPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  usePageTitle(t("processingActivitiesPage.title"));

  const organization = usePreloadedQuery<ProcessingActivityGraphListQuery>(
    processingActivitiesQuery,
    queryRef,
  );

  const paDocument = organization.node.processingActivitiesDocument;
  const paDefaultApproverIds = (paDocument?.defaultApprovers ?? []).map(
    a => a.id,
  );

  const goToDocument = (documentId: string) => {
    void navigate(
      `/organizations/${organizationId}/governance/documents/${documentId}`,
    );
  };

  const {
    data: activitiesData,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    ProcessingActivityGraphListQuery,
    ProcessingActivitiesPageFragment$key
  >(processingActivitiesPageFragment, organization.node);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ProcessingActivitiesConnectionKey,
  );
  const activities
    = activitiesData.processingActivities?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction = activities.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  const canPublishProcessingActivities
    = organization.node.canPublishProcessingActivities;

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("processingActivitiesPage.title")}
        description={t("processingActivitiesPage.description")}
      >
        {organization.node.canCreateProcessingActivity && (
          <CreateProcessingActivityDialog
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>
              {t("processingActivitiesPage.actions.add")}
            </Button>
          </CreateProcessingActivityDialog>
        )}
      </PageHeader>

      <div className="flex justify-end gap-2">
        {paDocument?.id && (
          <Button variant="secondary" asChild>
            <Link
              to={`/organizations/${organizationId}/governance/documents/${paDocument.id}`}
            >
              <IconPageTextLine size={16} />
              {t("processingActivitiesPage.actions.document")}
            </Link>
          </Button>
        )}
        {canPublishProcessingActivities && (
          <PublishProcessingActivityListDialog
            organizationId={organizationId}
            defaultApproverIds={paDefaultApproverIds}
            onPublished={goToDocument}
          >
            <Button variant="secondary" icon={IconUpload}>
              {t("processingActivitiesPage.actions.publish")}
            </Button>
          </PublishProcessingActivityListDialog>
        )}
      </div>

      {activities.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th className="px-3">
                      {t("processingActivitiesPage.columns.name")}
                    </Th>
                    <Th className="px-3">
                      {t("processingActivitiesPage.columns.purpose")}
                    </Th>
                    <Th className="px-3">
                      {t("processingActivitiesPage.columns.dataSubject")}
                    </Th>
                    <Th className="px-3">
                      {t("processingActivitiesPage.columns.lawfulBasis")}
                    </Th>
                    <Th className="px-3">
                      {t("processingActivitiesPage.columns.location")}
                    </Th>
                    <Th className="px-3">
                      {t("processingActivitiesPage.columns.internationalTransfers")}
                    </Th>
                    {hasAnyAction && (
                      <Th className="px-3">
                        {t("processingActivitiesPage.columns.actions")}
                      </Th>
                    )}
                  </Tr>
                </Thead>
                <Tbody>
                  {activities.map(activity => (
                    <ProcessingActivityListItem
                      key={activity.id}
                      processingActivityKey={activity}
                      connectionId={connectionId}
                      hasAnyAction={hasAnyAction}
                    />
                  ))}
                </Tbody>
              </Table>

              {hasNext && (
                <div className="p-4 border-t">
                  <Button
                    variant="secondary"
                    onClick={() => loadNext(10)}
                    disabled={isLoadingNext}
                  >
                    {isLoadingNext
                      ? t("processingActivitiesPage.actions.loading")
                      : t("processingActivitiesPage.actions.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {t("processingActivitiesPage.empty.title")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {t("processingActivitiesPage.empty.description")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
