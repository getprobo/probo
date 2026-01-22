import {
  downloadFile,
  promisifyMutation,
  sprintf,
  toDateInput,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  Dropdown,
  DropdownItem,
  IconArrowDown,
  IconChevronDown,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Spinner,
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
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";

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
import type { ProcessingActivityGraphListQuery } from "#/__generated__/core/ProcessingActivityGraphListQuery.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
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

interface ProcessingActivitiesPageProps {
  queryRef: PreloadedQuery<ProcessingActivityGraphListQuery>;
}

const processingActivitiesPageFragment = graphql`
    fragment ProcessingActivitiesPageFragment on Organization
    @refetchable(queryName: "ProcessingActivitiesPageRefetchQuery")
    @argumentDefinitions(
        first: { type: "Int", defaultValue: 10 }
        after: { type: "CursorKey" }
        snapshotId: { type: "ID", defaultValue: null }
    ) {
        id
        processingActivities(
            first: $first
            after: $after
            filter: { snapshotId: $snapshotId }
        )
            @connection(
                key: "ProcessingActivitiesPage_processingActivities"
                filters: ["filter"]
            ) {
            __id
            totalCount
            edges {
                node {
                    id
                    snapshotId
                    sourceId
                    name
                    purpose
                    dataSubjectCategory
                    personalDataCategory
                    lawfulBasis
                    location
                    internationalTransfers
                    createdAt
                    updatedAt
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
        snapshotId: { type: "ID", defaultValue: null }
    ) {
        id
        dataProtectionImpactAssessments(
            first: $first
            after: $after
            filter: { snapshotId: $snapshotId }
        )
            @connection(
                key: "ProcessingActivitiesPage_dataProtectionImpactAssessments"
                filters: ["filter"]
            ) {
            __id
            totalCount
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
                    createdAt
                    updatedAt
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
        snapshotId: { type: "ID", defaultValue: null }
    ) {
        id
        transferImpactAssessments(
            first: $first
            after: $after
            filter: { snapshotId: $snapshotId }
        )
            @connection(
                key: "ProcessingActivitiesPage_transferImpactAssessments"
                filters: ["filter"]
            ) {
            __id
            totalCount
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
                    createdAt
                    updatedAt
                }
            }
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }
`;

const exportProcessingActivitiesPDFMutation = graphql`
    mutation ProcessingActivitiesPageExportPDFMutation(
        $input: ExportProcessingActivitiesPDFInput!
    ) {
        exportProcessingActivitiesPDF(input: $input) {
            data
        }
    }
`;

const exportDataProtectionImpactAssessmentsPDFMutation = graphql`
    mutation ProcessingActivitiesPageExportDPIAPDFMutation(
        $input: ExportDataProtectionImpactAssessmentsPDFInput!
    ) {
        exportDataProtectionImpactAssessmentsPDF(input: $input) {
            data
        }
    }
`;

const exportTransferImpactAssessmentsPDFMutation = graphql`
    mutation ProcessingActivitiesPageExportTIAPDFMutation(
        $input: ExportTransferImpactAssessmentsPDFInput!
    ) {
        exportTransferImpactAssessmentsPDF(input: $input) {
            data
        }
    }
`;

export default function ProcessingActivitiesPage({
  queryRef,
}: ProcessingActivitiesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [activeTab, setActiveTab] = useState<"activities" | "dpia" | "tia">(
    "activities",
  );

  usePageTitle(__("Processing Activities"));

  const organization = usePreloadedQuery(processingActivitiesQuery, queryRef);

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
    { filter: { snapshotId: snapshotId || null } },
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

  const hasAnyAction
    = !isSnapshotMode
      && activities.some(({ canUpdate, canDelete }) => canUpdate || canDelete);

  const canExportPDF = organization.node.canExportProcessingActivities;

  const [exportPDF, isExportingPDF] = useMutationWithToasts<{
    response: {
      exportProcessingActivitiesPDF?: {
        data: string;
      };
    };
    variables: {
      input: {
        organizationId: string;
        filter: { snapshotId: string } | null;
      };
    };
  }>(exportProcessingActivitiesPDFMutation, {
    successMessage: __("PDF download started."),
    errorMessage: __("Failed to generate PDF"),
  });

  const handleExportPDF = async () => {
    await exportPDF({
      variables: {
        input: {
          organizationId: organizationId,
          filter: snapshotId ? { snapshotId } : null,
        },
      },
      onCompleted: (data) => {
        if (data.exportProcessingActivitiesPDF?.data) {
          downloadFile(
            data.exportProcessingActivitiesPDF.data,
            `processing-activities-${toDateInput(new Date().toISOString())}.pdf`,
          );
        }
      },
    });
  };

  const canExportDPIAPDF
    = organization.node.canExportDataProtectionImpactAssessments;

  const [exportDPIAPDF, isExportingDPIAPDF] = useMutationWithToasts<{
    response: {
      exportDataProtectionImpactAssessmentsPDF?: {
        data: string;
      };
    };
    variables: {
      input: {
        organizationId: string;
        filter: { snapshotId: string } | null;
      };
    };
  }>(exportDataProtectionImpactAssessmentsPDFMutation, {
    successMessage: __("PDF download started."),
    errorMessage: __("Failed to generate PDF"),
  });

  const handleExportDPIAPDF = async () => {
    await exportDPIAPDF({
      variables: {
        input: {
          organizationId: organizationId,
          filter: snapshotId ? { snapshotId } : null,
        },
      },
      onCompleted: (data) => {
        if (data.exportDataProtectionImpactAssessmentsPDF?.data) {
          downloadFile(
            data.exportDataProtectionImpactAssessmentsPDF.data,
            `data-protection-impact-assessments-${toDateInput(new Date().toISOString())}.pdf`,
          );
        }
      },
    });
  };

  const canExportTIAPDF
    = organization.node.canExportTransferImpactAssessments;

  const [exportTIAPDF, isExportingTIAPDF] = useMutationWithToasts<{
    response: {
      exportTransferImpactAssessmentsPDF?: {
        data: string;
      };
    };
    variables: {
      input: {
        organizationId: string;
        filter: { snapshotId: string } | null;
      };
    };
  }>(exportTransferImpactAssessmentsPDFMutation, {
    successMessage: __("PDF download started."),
    errorMessage: __("Failed to generate PDF"),
  });

  const handleExportTIAPDF = async () => {
    await exportTIAPDF({
      variables: {
        input: {
          organizationId: organizationId,
          filter: snapshotId ? { snapshotId } : null,
        },
      },
      onCompleted: (data) => {
        if (data.exportTransferImpactAssessmentsPDF?.data) {
          downloadFile(
            data.exportTransferImpactAssessmentsPDF.data,
            `transfer-impact-assessments-${toDateInput(new Date().toISOString())}.pdf`,
          );
        }
      },
    });
  };

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <PageHeader
        title={__("Processing Activities")}
        description={__("Manage your processing activities under GDPR")}
      >
        {(canExportPDF || canExportDPIAPDF || canExportTIAPDF) && (
          <Dropdown
            toggle={(
              <Button
                variant="secondary"
                icon={IconArrowDown}
                iconAfter={IconChevronDown}
              >
                {__("Export")}
              </Button>
            )}
          >
            {canExportPDF && (
              <DropdownItem
                onClick={() => void handleExportPDF()}
                disabled={isExportingPDF}
                icon={isExportingPDF ? Spinner : undefined}
              >
                {__("Processing Activities")}
              </DropdownItem>
            )}
            {canExportDPIAPDF && (
              <DropdownItem
                onClick={() => void handleExportDPIAPDF()}
                disabled={isExportingDPIAPDF}
                icon={isExportingDPIAPDF ? Spinner : undefined}
              >
                {__("Data Protection Impact Assessments")}
              </DropdownItem>
            )}
            {canExportTIAPDF && (
              <DropdownItem
                onClick={() => void handleExportTIAPDF()}
                disabled={isExportingTIAPDF}
                icon={isExportingTIAPDF ? Spinner : undefined}
              >
                {__("Transfer Impact Assessments")}
              </DropdownItem>
            )}
          </Dropdown>
        )}
        {!isSnapshotMode
          && activeTab === "activities"
          && organization.node.canCreateProcessingActivity && (
          <CreateProcessingActivityDialog
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>
              {__("Add processing activity")}
            </Button>
          </CreateProcessingActivityDialog>
        )}
      </PageHeader>

      <Tabs>
        <TabItem
          active={activeTab === "activities"}
          onClick={() => setActiveTab("activities")}
        >
          {__("Processing Activities")}
        </TabItem>
        <TabItem
          active={activeTab === "dpia"}
          onClick={() => setActiveTab("dpia")}
        >
          {__("Data Protection Impact Assessments")}
        </TabItem>
        <TabItem
          active={activeTab === "tia"}
          onClick={() => setActiveTab("tia")}
        >
          {__("Transfer Impact Assessments")}
        </TabItem>
      </Tabs>

      {activeTab === "activities" && (
        <>
          {activities.length > 0
            ? (
                <Card>
                  <Table>
                    <Thead>
                      <Tr>
                        <Th className="px-3">{__("Name")}</Th>
                        <Th className="px-3">
                          {__("Purpose")}
                        </Th>
                        <Th className="px-3">
                          {__("Data Subject")}
                        </Th>
                        <Th className="px-3">
                          {__("Lawful Basis")}
                        </Th>
                        <Th className="px-3">
                          {__("Location")}
                        </Th>
                        <Th className="px-3">
                          {__("International Transfers")}
                        </Th>
                        {hasAnyAction && (
                          <Th className="px-3">
                            {__("Actions")}
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
                          ? __("Loading...")
                          : __("Load more")}
                      </Button>
                    </div>
                  )}
                </Card>
              )
            : (
                <Card padded>
                  <div className="text-center py-12">
                    <h3 className="text-lg font-semibold mb-2">
                      {__("No processing activities yet")}
                    </h3>
                    <p className="text-txt-tertiary mb-4">
                      {__(
                        "Create your first processing activity to get started with GDPR compliance.",
                      )}
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
                        <Th>{__("Processing Activity")}</Th>
                        <Th>{__("Description")}</Th>
                        <Th>{__("Potential Risk")}</Th>
                        <Th>{__("Residual Risk")}</Th>
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
                          ? __("Loading...")
                          : __("Load more")}
                      </Button>
                    </div>
                  )}
                </Card>
              )
            : (
                <Card padded>
                  <div className="text-center py-12">
                    <h3 className="text-lg font-semibold mb-2">
                      {__(
                        "No Data Protection Impact Assessments yet",
                      )}
                    </h3>
                    <p className="text-txt-tertiary mb-4">
                      {__(
                        "DPIAs are created from within individual processing activities.",
                      )}
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
                        <Th>{__("Processing Activity")}</Th>
                        <Th>{__("Data Subjects")}</Th>
                        <Th>{__("Transfer")}</Th>
                        <Th>{__("Local Law Risk")}</Th>
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
                          ? __("Loading...")
                          : __("Load more")}
                      </Button>
                    </div>
                  )}
                </Card>
              )
            : (
                <Card padded>
                  <div className="text-center py-12">
                    <h3 className="text-lg font-semibold mb-2">
                      {__("No Transfer Impact Assessments yet")}
                    </h3>
                    <p className="text-txt-tertiary mb-4">
                      {__(
                        "TIAs are created from within individual processing activities.",
                      )}
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
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [deleteActivity] = useMutation(deleteProcessingActivityMutation);
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
        message: sprintf(
          __(
            "This will permanently delete the processing activity %s. This action cannot be undone.",
          ),
          activity.name,
        ),
      },
    );
  };

  const activityUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/processing-activities/${activity.id}`
      : `/organizations/${organizationId}/processing-activities/${activity.id}`;

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
      <Td>{getLawfulBasisLabel(activity.lawfulBasis, __)}</Td>
      <Td>{activity.location || "-"}</Td>
      <Td>
        <Badge
          variant={
            activity.internationalTransfers ? "warning" : "success"
          }
        >
          {activity.internationalTransfers ? __("Yes") : __("No")}
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
                {__("Delete")}
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
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const activityUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/processing-activities/${dpia.processingActivity.id}#dpia`
      : `/organizations/${organizationId}/processing-activities/${dpia.processingActivity.id}#dpia`;

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
                {getResidualRiskLabel(dpia.residualRisk, __)}
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
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const activityUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/processing-activities/${tia.processingActivity.id}#tia`
      : `/organizations/${organizationId}/processing-activities/${tia.processingActivity.id}#tia`;

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
