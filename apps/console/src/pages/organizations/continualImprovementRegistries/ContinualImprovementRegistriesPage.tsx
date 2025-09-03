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
import { CreateContinualImprovementRegistryDialog } from "./dialogs/CreateContinualImprovementRegistryDialog";
import { deleteContinualImprovementRegistryMutation, ContinualImprovementRegistriesConnectionKey } from "../../../hooks/graph/ContinualImprovementRegistryGraph";
import { sprintf, promisifyMutation, getStatusVariant, getStatusLabel } from "@probo/helpers";
import { SnapshotBanner } from "/components/SnapshotBanner";
import type { NodeOf } from "/types";
import type { ContinualImprovementRegistriesPageQuery } from "./__generated__/ContinualImprovementRegistriesPageQuery.graphql";
import type {
  ContinualImprovementRegistriesPageFragment$key,
  ContinualImprovementRegistriesPageFragment$data,
} from "./__generated__/ContinualImprovementRegistriesPageFragment.graphql";

interface ContinualImprovementRegistriesPageProps {
  queryRef: PreloadedQuery<ContinualImprovementRegistriesPageQuery>;
}

const continualImprovementRegistriesPageFragment = graphql`
  fragment ContinualImprovementRegistriesPageFragment on Organization
  @refetchable(queryName: "ContinualImprovementRegistriesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    id
    continualImprovementRegistries(
      first: $first
      after: $after
      filter: { snapshotId: $snapshotId }
    )
      @connection(key: "ContinualImprovementRegistriesPage_continualImprovementRegistries", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          id
          snapshotId
          sourceId
          referenceId
          description
          source
          targetDate
          status
          priority
          owner {
            id
            fullName
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

export default function ContinualImprovementRegistriesPage({ queryRef }: ContinualImprovementRegistriesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("Continual Improvements"));

  const organization = usePreloadedQuery(
    graphql`
      query ContinualImprovementRegistriesPageQuery($organizationId: ID!, $snapshotId: ID) {
        node(id: $organizationId) {
          ... on Organization {
            ...ContinualImprovementRegistriesPageFragment @arguments(snapshotId: $snapshotId)
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
    ContinualImprovementRegistriesPageQuery,
    ContinualImprovementRegistriesPageFragment$key
  >(continualImprovementRegistriesPageFragment, organization.node);
  if (!data) {
    return <div>{__("Organization not found")}</div>;
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ContinualImprovementRegistriesConnectionKey,
    { filter: { snapshotId: snapshotId || null } }
  );
  const registries = data?.continualImprovementRegistries?.edges?.map((edge) => edge.node) ?? [];

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <PageHeader title={__("Continual Improvements")} description={__("Manage your continual improvement entries")}>
        {!isSnapshotMode && (
          <CreateContinualImprovementRegistryDialog
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>
              {__("Add continual improvement")}
            </Button>
          </CreateContinualImprovementRegistryDialog>
        )}
      </PageHeader>

      {registries.length > 0 ? (
        <Card>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Reference ID")}</Th>
                <Th>{__("Description")}</Th>
                <Th>{__("Status")}</Th>
                <Th>{__("Priority")}</Th>
                <Th>{__("Owner")}</Th>
                <Th>{__("Target Date")}</Th>
                {!isSnapshotMode && <Th>{__("Actions")}</Th>}
              </Tr>
            </Thead>
            <Tbody>
              {registries.map((registry) => (
                <RegistryRow
                  key={registry.id}
                  registry={registry}
                  connectionId={connectionId}
                  snapshotId={snapshotId}
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
              {__("No continual improvement entries yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first continual improvement entry to get started.")}
            </p>
          </div>
        </Card>
      )}
    </div>
  );
}

function RegistryRow({
  registry,
  connectionId,
  snapshotId,
}: {
  registry: NodeOf<NonNullable<ContinualImprovementRegistriesPageFragment$data['continualImprovementRegistries']>>;
  connectionId: string;
  snapshotId?: string;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const [deleteRegistry] = useMutation(deleteContinualImprovementRegistryMutation);
  const confirm = useConfirm();
  const isSnapshotMode = Boolean(snapshotId);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteRegistry)({
          variables: {
            input: {
              continualImprovementRegistryId: registry.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the continual improvement registry entry %s. This action cannot be undone."
          ),
          registry.referenceId
        ),
      }
    );
  };

  const detailsUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/continual-improvement-registries/${registry.id}`
    : `/organizations/${organizationId}/continual-improvement-registries/${registry.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>
        <span className="font-mono text-sm">{registry.referenceId}</span>
      </Td>
      <Td>{registry.description || "-"}</Td>
      <Td>
        <Badge variant={getStatusVariant(registry.status)}>
          {getStatusLabel(registry.status)}
        </Badge>
      </Td>
      <Td>
        <Badge variant={registry.priority === "HIGH" ? "danger" : registry.priority === "MEDIUM" ? "warning" : "success"}>
          {registry.priority === "HIGH" ? __("High") : registry.priority === "MEDIUM" ? __("Medium") : __("Low")}
        </Badge>
      </Td>
      <Td>{registry.owner?.fullName || "-"}</Td>
      <Td>
        {registry.targetDate ? (
          <time dateTime={registry.targetDate}>
            {formatDate(registry.targetDate)}
          </time>
        ) : (
          <span className="text-txt-tertiary">{__("No target date")}</span>
        )}
      </Td>
      {!isSnapshotMode && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            <DropdownItem
              icon={IconTrashCan}
              variant="danger"
              onSelect={handleDelete}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
