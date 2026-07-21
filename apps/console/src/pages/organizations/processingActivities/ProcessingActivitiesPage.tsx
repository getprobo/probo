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

import { promisifyMutation } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  DropdownItem,
  IconPageTextLine,
  IconPlusLarge,
  IconTrashCan,
  IconUpload,
  PageHeader,
  TabItem,
  Table,
  Tabs,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";

import type {
  ProcessingActivitiesPageDPIAFragment$data,
  ProcessingActivitiesPageDPIAFragment$key,
} from "#/__generated__/core/ProcessingActivitiesPageDPIAFragment.graphql";
import type {
  ProcessingActivitiesPageFragment$data,
  ProcessingActivitiesPageFragment$key,
} from "#/__generated__/core/ProcessingActivitiesPageFragment.graphql";
import type {
  ProcessingActivitiesPageTIAFragment$data,
  ProcessingActivitiesPageTIAFragment$key,
} from "#/__generated__/core/ProcessingActivitiesPageTIAFragment.graphql";
import type { ProcessingActivityGraphDeleteMutation } from "#/__generated__/core/ProcessingActivityGraphDeleteMutation.graphql";
import type { ProcessingActivityGraphListQuery } from "#/__generated__/core/ProcessingActivityGraphListQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import {
  getLawfulBasisLabel,
  getResidualRiskLabel,
} from "../../../components/form/ProcessingActivityEnumOptions";
import {
  deleteProcessingActivityMutation,
  ProcessingActivitiesConnectionKey,
  processingActivitiesQuery,
} from "../../../hooks/graph/ProcessingActivityGraph";

import { CreateProcessingActivityDialog } from "./dialogs/CreateProcessingActivityDialog";
import { PublishDataProtectionImpactAssessmentListDialog } from "./dialogs/PublishDataProtectionImpactAssessmentListDialog";
import { PublishProcessingActivityListDialog } from "./dialogs/PublishProcessingActivityListDialog";
import { PublishTransferImpactAssessmentListDialog } from "./dialogs/PublishTransferImpactAssessmentListDialog";

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
            @connection(
                key: "ProcessingActivitiesPage_processingActivities"
            ) {
            edges {
                node {
                    id
                    name
                    purpose
                    dataSubjectCategory
                    lawfulBasis
                    location
                    internationalTransfers
                    canUpdate: permission(
                        action: "core:processing-activity:update"
                    )
                    canDelete: permission(
                        action: "core:processing-activity:delete"
                    )
                }
            }
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }
`;

const dpiaListPageFragment = graphql`
    fragment ProcessingActivitiesPageDPIAFragment on Organization
    @refetchable(queryName: "ProcessingActivitiesPageDPIARefetchQuery")
    @argumentDefinitions(
        first: { type: "Int", defaultValue: 10 }
        after: { type: "CursorKey" }
    ) {
        id
        dataProtectionImpactAssessments(first: $first, after: $after)
            @connection(
                key: "ProcessingActivitiesPage_dataProtectionImpactAssessments"
            ) {
            edges {
                node {
                    id
                    description
                    potentialRisk
                    residualRisk
                    processingActivity {
                        id
                        name
                    }
                }
            }
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }
`;

const tiaListPageFragment = graphql`
    fragment ProcessingActivitiesPageTIAFragment on Organization
    @refetchable(queryName: "ProcessingActivitiesPageTIARefetchQuery")
    @argumentDefinitions(
        first: { type: "Int", defaultValue: 10 }
        after: { type: "CursorKey" }
    ) {
        id
        transferImpactAssessments(first: $first, after: $after)
            @connection(
                key: "ProcessingActivitiesPage_transferImpactAssessments"
            ) {
            edges {
                node {
                    id
                    dataSubjects
                    transfer
                    localLawRisk
                    processingActivity {
                        id
                        name
                    }
                }
            }
            pageInfo {
                hasNextPage
                endCursor
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
  const [activeTab, setActiveTab] = useState<"activities" | "dpia" | "tia">(
    "activities",
  );

  usePageTitle(t("processingActivitiesPage.title"));

  const organization = usePreloadedQuery<ProcessingActivityGraphListQuery>(processingActivitiesQuery, queryRef);

  const paDocument = organization.node.processingActivitiesDocument;
  const dpiaDocument = organization.node.dataProtectionImpactAssessmentsDocument;
  const tiaDocument = organization.node.transferImpactAssessmentsDocument;
  const paDefaultApproverIds = (paDocument?.defaultApprovers ?? []).map(a => a.id);
  const dpiaDefaultApproverIds = (dpiaDocument?.defaultApprovers ?? []).map(a => a.id);
  const tiaDefaultApproverIds = (tiaDocument?.defaultApprovers ?? []).map(a => a.id);

  const goToDocument = (documentId: string) => {
    void navigate(`/organizations/${organizationId}/documents/${documentId}`);
  };

  const {
    data: activitiesData,
    loadNext: loadNextActivities,
    hasNext: hasNextActivities,
    isLoadingNext: isLoadingNextActivities,
  } = usePaginationFragment<
    ProcessingActivityGraphListQuery,
    ProcessingActivitiesPageFragment$key
  >(processingActivitiesPageFragment, organization.node);

  const {
    data: dpiaData,
    loadNext: loadNextDPIAs,
    hasNext: hasNextDPIAs,
    isLoadingNext: isLoadingNextDPIAs,
  } = usePaginationFragment<
    ProcessingActivityGraphListQuery,
    ProcessingActivitiesPageDPIAFragment$key
  >(dpiaListPageFragment, organization.node);

  const {
    data: tiaData,
    loadNext: loadNextTIAs,
    hasNext: hasNextTIAs,
    isLoadingNext: isLoadingNextTIAs,
  } = usePaginationFragment<
    ProcessingActivityGraphListQuery,
    ProcessingActivitiesPageTIAFragment$key
  >(tiaListPageFragment, organization.node);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ProcessingActivitiesConnectionKey,
  );
  const activities
    = activitiesData?.processingActivities?.edges?.map(edge => edge.node)
      ?? [];
  const dpias
    = dpiaData?.dataProtectionImpactAssessments?.edges?.map(
      edge => edge.node,
    ) ?? [];
  const tias
    = tiaData?.transferImpactAssessments?.edges?.map(edge => edge.node)
      ?? [];

  const hasAnyAction = activities.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  const canPublishProcessingActivities
    = organization.node.canPublishProcessingActivities;
  const canPublishDPIA
    = organization.node.canPublishDataProtectionImpactAssessments;
  const canPublishTIA
    = organization.node.canPublishTransferImpactAssessments;

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("processingActivitiesPage.title")}
        description={t("processingActivitiesPage.description")}
      >
        {activeTab === "activities"
          && organization.node.canCreateProcessingActivity && (
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

      <Tabs>
        <TabItem
          active={activeTab === "activities"}
          onClick={() => setActiveTab("activities")}
        >
          {t("processingActivitiesPage.tabs.activities")}
        </TabItem>
        <TabItem
          active={activeTab === "dpia"}
          onClick={() => setActiveTab("dpia")}
        >
          {t("processingActivitiesPage.tabs.dpia")}
        </TabItem>
        <TabItem
          active={activeTab === "tia"}
          onClick={() => setActiveTab("tia")}
        >
          {t("processingActivitiesPage.tabs.tia")}
        </TabItem>
      </Tabs>

      <div className="flex justify-end gap-2">
        {activeTab === "activities" && (
          <>
            {paDocument?.id && (
              <Button variant="secondary" asChild>
                <Link
                  to={`/organizations/${organizationId}/documents/${paDocument.id}`}
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
          </>
        )}
        {activeTab === "dpia" && (
          <>
            {dpiaDocument?.id && (
              <Button variant="secondary" asChild>
                <Link
                  to={`/organizations/${organizationId}/documents/${dpiaDocument.id}`}
                >
                  <IconPageTextLine size={16} />
                  {t("processingActivitiesPage.actions.document")}
                </Link>
              </Button>
            )}
            {canPublishDPIA && (
              <PublishDataProtectionImpactAssessmentListDialog
                organizationId={organizationId}
                defaultApproverIds={dpiaDefaultApproverIds}
                onPublished={goToDocument}
              >
                <Button variant="secondary" icon={IconUpload}>
                  {t("processingActivitiesPage.actions.publish")}
                </Button>
              </PublishDataProtectionImpactAssessmentListDialog>
            )}
          </>
        )}
        {activeTab === "tia" && (
          <>
            {tiaDocument?.id && (
              <Button variant="secondary" asChild>
                <Link
                  to={`/organizations/${organizationId}/documents/${tiaDocument.id}`}
                >
                  <IconPageTextLine size={16} />
                  {t("processingActivitiesPage.actions.document")}
                </Link>
              </Button>
            )}
            {canPublishTIA && (
              <PublishTransferImpactAssessmentListDialog
                organizationId={organizationId}
                defaultApproverIds={tiaDefaultApproverIds}
                onPublished={goToDocument}
              >
                <Button variant="secondary" icon={IconUpload}>
                  {t("processingActivitiesPage.actions.publish")}
                </Button>
              </PublishTransferImpactAssessmentListDialog>
            )}
          </>
        )}
      </div>

      {activeTab === "activities" && (
        <>
          {activities.length > 0
            ? (
                <Card>
                  <Table>
                    <Thead>
                      <Tr>
                        <Th className="px-3">{t("processingActivitiesPage.columns.name")}</Th>
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
                        <ActivityRow
                          key={activity.id}
                          activity={activity}
                          connectionId={connectionId}
                          hasAnyAction={hasAnyAction}
                        />
                      ))}
                    </Tbody>
                  </Table>

                  {hasNextActivities && (
                    <div className="p-4 border-t">
                      <Button
                        variant="secondary"
                        onClick={() => loadNextActivities(10)}
                        disabled={isLoadingNextActivities}
                      >
                        {isLoadingNextActivities
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
                      {t("processingActivitiesPage.empty.activitiesTitle")}
                    </h3>
                    <p className="text-txt-tertiary mb-4">
                      {t("processingActivitiesPage.empty.activitiesDescription")}
                    </p>
                  </div>
                </Card>
              )}
        </>
      )}

      {activeTab === "dpia" && (
        <>
          {dpias.length > 0
            ? (
                <Card>
                  <Table>
                    <Thead>
                      <Tr>
                        <Th>{t("processingActivitiesPage.columns.processingActivity")}</Th>
                        <Th>{t("processingActivitiesPage.columns.description")}</Th>
                        <Th>{t("processingActivitiesPage.columns.potentialRisk")}</Th>
                        <Th>{t("processingActivitiesPage.columns.residualRisk")}</Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {dpias.map(dpia => (
                        <DPIARow key={dpia.id} dpia={dpia} />
                      ))}
                    </Tbody>
                  </Table>

                  {hasNextDPIAs && (
                    <div className="p-4 border-t">
                      <Button
                        variant="secondary"
                        onClick={() => loadNextDPIAs(10)}
                        disabled={isLoadingNextDPIAs}
                      >
                        {isLoadingNextDPIAs
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
                      {t("processingActivitiesPage.empty.dpiaTitle")}
                    </h3>
                    <p className="text-txt-tertiary mb-4">
                      {t("processingActivitiesPage.empty.dpiaDescription")}
                    </p>
                  </div>
                </Card>
              )}
        </>
      )}

      {activeTab === "tia" && (
        <>
          {tias.length > 0
            ? (
                <Card>
                  <Table>
                    <Thead>
                      <Tr>
                        <Th>{t("processingActivitiesPage.columns.processingActivity")}</Th>
                        <Th>{t("processingActivitiesPage.columns.dataSubjects")}</Th>
                        <Th>{t("processingActivitiesPage.columns.transfer")}</Th>
                        <Th>{t("processingActivitiesPage.columns.localLawRisk")}</Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {tias.map(tia => (
                        <TIARow key={tia.id} tia={tia} />
                      ))}
                    </Tbody>
                  </Table>

                  {hasNextTIAs && (
                    <div className="p-4 border-t">
                      <Button
                        variant="secondary"
                        onClick={() => loadNextTIAs(10)}
                        disabled={isLoadingNextTIAs}
                      >
                        {isLoadingNextTIAs
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
                      {t("processingActivitiesPage.empty.tiaTitle")}
                    </h3>
                    <p className="text-txt-tertiary mb-4">
                      {t("processingActivitiesPage.empty.tiaDescription")}
                    </p>
                  </div>
                </Card>
              )}
        </>
      )}
    </div>
  );
}

function ActivityRow({
  activity,
  connectionId,
  hasAnyAction,
}: {
  activity: NodeOf<
    NonNullable<
      ProcessingActivitiesPageFragment$data["processingActivities"]
    >
  >;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const [deleteActivity] = useMutation<ProcessingActivityGraphDeleteMutation>(deleteProcessingActivityMutation);
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteActivity)({
          variables: {
            input: {
              processingActivityId: activity.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: t("processingActivitiesPage.deleteConfirmation", { name: activity.name }),
      },
    );
  };

  const activityUrl
    = `/organizations/${organizationId}/processing-activities/${activity.id}`;

  return (
    <Tr to={activityUrl}>
      <Td>
        <span className="font-semibold">{activity.name}</span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {activity.purpose || "-"}
        </span>
      </Td>
      <Td>{activity.dataSubjectCategory || "-"}</Td>
      <Td>{getLawfulBasisLabel(activity.lawfulBasis, t)}</Td>
      <Td>{activity.location || "-"}</Td>
      <Td>
        <Badge
          variant={
            activity.internationalTransfers ? "warning" : "success"
          }
        >
          {activity.internationalTransfers ? t("processingActivitiesPage.answers.yes") : t("processingActivitiesPage.answers.no")}
        </Badge>
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {activity.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {t("processingActivitiesPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}

function DPIARow({
  dpia,
}: {
  dpia: NodeOf<
    NonNullable<
      ProcessingActivitiesPageDPIAFragment$data["dataProtectionImpactAssessments"]
    >
  >;
}) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation();

  const activityUrl
    = `/organizations/${organizationId}/processing-activities/${dpia.processingActivity.id}#dpia`;

  return (
    <Tr to={activityUrl}>
      <Td>
        <span className="font-semibold">
          {dpia.processingActivity.name}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary line-clamp-2">
          {dpia.description || "-"}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary line-clamp-2">
          {dpia.potentialRisk || "-"}
        </span>
      </Td>
      <Td>
        {dpia.residualRisk
          ? (
              <Badge
                variant={
                  dpia.residualRisk === "LOW"
                    ? "success"
                    : dpia.residualRisk === "MEDIUM"
                      ? "warning"
                      : "danger"
                }
              >
                {getResidualRiskLabel(dpia.residualRisk, t)}
              </Badge>
            )
          : (
              "-"
            )}
      </Td>
    </Tr>
  );
}

function TIARow({
  tia,
}: {
  tia: NodeOf<
    NonNullable<
      ProcessingActivitiesPageTIAFragment$data["transferImpactAssessments"]
    >
  >;
}) {
  const organizationId = useOrganizationId();

  const activityUrl
    = `/organizations/${organizationId}/processing-activities/${tia.processingActivity.id}#tia`;

  return (
    <Tr to={activityUrl}>
      <Td>
        <span className="font-semibold">
          {tia.processingActivity.name}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary line-clamp-2">
          {tia.dataSubjects || "-"}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary line-clamp-2">
          {tia.transfer || "-"}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary line-clamp-2">
          {tia.localLawRisk || "-"}
        </span>
      </Td>
    </Tr>
  );
}
