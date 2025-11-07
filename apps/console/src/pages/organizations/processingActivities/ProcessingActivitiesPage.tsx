import {
  Button,
  IconPlusLarge,
  PageHeader,
  Card,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  Table,
  useConfirm,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import { getLawfulBasisLabel } from "../../../components/form/ProcessingActivityEnumOptions";
import {
  ConnectionHandler,
  graphql,
  usePaginationFragment,
  usePreloadedQuery,
  useMutation,
  type PreloadedQuery,
} from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useParams } from "react-router";
import { CreateProcessingActivityDialog } from "./dialogs/CreateProcessingActivityDialog";
import { deleteProcessingActivityMutation, ProcessingActivitiesConnectionKey } from "../../../hooks/graph/ProcessingActivityGraph";
import { sprintf, promisifyMutation } from "@probo/helpers";
import { SnapshotBanner } from "/components/SnapshotBanner";
import { Authorized } from "/permissions";
import { isAuthorized } from "/permissions";
import type { NodeOf } from "/types";
import type { ProcessingActivitiesPageQuery } from "./__generated__/ProcessingActivitiesPageQuery.graphql";
import type {
  ProcessingActivitiesPageFragment$key,
  ProcessingActivitiesPageFragment$data,
} from "./__generated__/ProcessingActivitiesPageFragment.graphql";

interface ProcessingActivitiesPageProps {
  queryRef: PreloadedQuery<ProcessingActivitiesPageQuery>;
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
      @connection(key: "ProcessingActivitiesPage_processingActivities", filters: ["filter"]) {
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
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export default function ProcessingActivitiesPage({ queryRef }: ProcessingActivitiesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);


  usePageTitle(__("Processing Activities"));

  const organization = usePreloadedQuery(
    graphql`
      query ProcessingActivitiesPageQuery($organizationId: ID!, $snapshotId: ID) {
        node(id: $organizationId) {
          ... on Organization {
            ...ProcessingActivitiesPageFragment @arguments(snapshotId: $snapshotId)
          }
        }
      }
    `,
    queryRef
  );

  const {
    data,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    ProcessingActivitiesPageQuery,
    ProcessingActivitiesPageFragment$key
  >(processingActivitiesPageFragment, organization.node);
  if (!data) {
    return <div>{__("Organization not found")}</div>;
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ProcessingActivitiesConnectionKey,
    { filter: { snapshotId: snapshotId || null } }
  );
  const activities = data?.processingActivities?.edges?.map((edge) => edge.node) ?? [];

  const hasAnyAction = !isSnapshotMode && (
    isAuthorized(organizationId, "ProcessingActivity", "updateProcessingActivity") ||
    isAuthorized(organizationId, "ProcessingActivity", "deleteProcessingActivity")
  );

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <PageHeader title={__("Processing Activities")} description={__("Manage your processing activities under GDPR")}>
        {!isSnapshotMode && (
          <Authorized entity="Organization" action="createProcessingActivity">
            <CreateProcessingActivityDialog
              organizationId={organizationId}
              connectionId={connectionId}
            >
              <Button icon={IconPlusLarge}>
                {__("Add processing activity")}
              </Button>
            </CreateProcessingActivityDialog>
          </Authorized>
        )}
      </PageHeader>

      {activities.length > 0 ? (
        <Card>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Name")}</Th>
                <Th>{__("Purpose")}</Th>
                <Th>{__("Data Subject")}</Th>
                <Th>{__("Lawful Basis")}</Th>
                <Th>{__("Location")}</Th>
                <Th>{__("International Transfers")}</Th>
                {hasAnyAction && <Th>{__("Actions")}</Th>}
              </Tr>
            </Thead>
            <Tbody>
              {activities.map((activity) => (
                <ActivityRow
                  key={activity.id}
                  activity={activity}
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
                {isLoadingNext ? __("Loading...") : __("Load more")}
              </Button>
            </div>
          )}
        </Card>
      ) : (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No processing activities yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first processing activity to get started with GDPR compliance.")}
            </p>
          </div>
        </Card>
      )}
    </div>
  );
}

function ActivityRow({
  activity,
  connectionId,
  hasAnyAction,
}: {
  activity: NodeOf<NonNullable<ProcessingActivitiesPageFragment$data['processingActivities']>>;
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
            "This will permanently delete the processing activity %s. This action cannot be undone."
          ),
          activity.name
        ),
      }
    );
  };

  const activityUrl = isSnapshotMode && snapshotId
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
        <Badge variant={activity.internationalTransfers ? "warning" : "success"}>
          {activity.internationalTransfers ? __("Yes") : __("No")}
        </Badge>
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            <Authorized entity="ProcessingActivity" action="deleteProcessingActivity">
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {__("Delete")}
              </DropdownItem>
            </Authorized>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
