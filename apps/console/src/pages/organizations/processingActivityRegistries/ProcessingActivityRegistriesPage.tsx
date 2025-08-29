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
import { getLawfulBasisLabel } from "../../../components/form/ProcessingActivityRegistryEnumOptions";
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
import { CreateProcessingActivityRegistryDialog } from "./dialogs/CreateProcessingActivityRegistryDialog";
import { deleteProcessingActivityRegistryMutation, ProcessingActivityRegistriesConnectionKey } from "../../../hooks/graph/ProcessingActivityRegistryGraph";
import { sprintf, promisifyMutation } from "@probo/helpers";
import { SnapshotBanner } from "/components/SnapshotBanner";
import type { NodeOf } from "/types";
import type { ProcessingActivityRegistriesPageQuery } from "./__generated__/ProcessingActivityRegistriesPageQuery.graphql";
import type {
  ProcessingActivityRegistriesPageFragment$key,
  ProcessingActivityRegistriesPageFragment$data,
} from "./__generated__/ProcessingActivityRegistriesPageFragment.graphql";

interface ProcessingActivityRegistriesPageProps {
  queryRef: PreloadedQuery<ProcessingActivityRegistriesPageQuery>;
}

const processingActivityRegistriesPageFragment = graphql`
  fragment ProcessingActivityRegistriesPageFragment on Organization
  @refetchable(queryName: "ProcessingActivityRegistriesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    id
    processingActivityRegistries(
      first: $first
      after: $after
      filter: { snapshotId: $snapshotId }
    )
      @connection(key: "ProcessingActivityRegistriesPage_processingActivityRegistries", filters: ["filter"]) {
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

export default function ProcessingActivityRegistriesPage({ queryRef }: ProcessingActivityRegistriesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);


  usePageTitle(__("Processing Activity Registries"));

  const organization = usePreloadedQuery(
    graphql`
      query ProcessingActivityRegistriesPageQuery($organizationId: ID!, $snapshotId: ID) {
        node(id: $organizationId) {
          ... on Organization {
            ...ProcessingActivityRegistriesPageFragment @arguments(snapshotId: $snapshotId)
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
    ProcessingActivityRegistriesPageQuery,
    ProcessingActivityRegistriesPageFragment$key
  >(processingActivityRegistriesPageFragment, organization.node);
  if (!data) {
    return <div>{__("Organization not found")}</div>;
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ProcessingActivityRegistriesConnectionKey,
    { filter: { snapshotId: snapshotId || null } }
  );
  const registries = data?.processingActivityRegistries?.edges?.map((edge) => edge.node) ?? [];

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <PageHeader title={__("Processing Activity Registries")} description={__("Manage your processing activity registry entries under GDPR")}>
        {!isSnapshotMode && (
          <CreateProcessingActivityRegistryDialog
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>
              {__("Add processing activity registry")}
            </Button>
          </CreateProcessingActivityRegistryDialog>
        )}
      </PageHeader>

      {registries.length > 0 ? (
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
                {!isSnapshotMode && <Th>{__("Actions")}</Th>}
              </Tr>
            </Thead>
            <Tbody>
              {registries.map((registry) => (
                <RegistryRow
                  key={registry.id}
                  registry={registry}
                  connectionId={connectionId}
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
              {__("No processing activity registry entries yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first processing activity registry entry to get started with GDPR compliance.")}
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
}: {
  registry: NodeOf<NonNullable<ProcessingActivityRegistriesPageFragment$data['processingActivityRegistries']>>;
  connectionId: string;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [deleteRegistry] = useMutation(deleteProcessingActivityRegistryMutation);
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteRegistry)({
          variables: {
            input: {
              processingActivityRegistryId: registry.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the processing activity registry entry %s. This action cannot be undone."
          ),
          registry.name
        ),
      }
    );
  };

  const registryUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/processing-activity-registries/${registry.id}`
    : `/organizations/${organizationId}/processing-activity-registries/${registry.id}`;

  return (
    <Tr to={registryUrl}>
      <Td>
        <span className="font-semibold">{registry.name}</span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {registry.purpose || "-"}
        </span>
      </Td>
      <Td>{registry.dataSubjectCategory || "-"}</Td>
      <Td>{getLawfulBasisLabel(registry.lawfulBasis, __)}</Td>
      <Td>{registry.location || "-"}</Td>
      <Td>
        <Badge variant={registry.internationalTransfers ? "warning" : "success"}>
          {registry.internationalTransfers ? __("Yes") : __("No")}
        </Badge>
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
