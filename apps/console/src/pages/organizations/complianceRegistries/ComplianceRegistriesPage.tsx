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
  graphql,
  usePaginationFragment,
  usePreloadedQuery,
  useMutation,
  type PreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { CreateComplianceRegistryDialog } from "./dialogs/CreateComplianceRegistryDialog";
import { deleteComplianceRegistryMutation } from "../../../hooks/graph/ComplianceRegistryGraph";
import { sprintf, promisifyMutation, getStatusVariant, getStatusLabel } from "@probo/helpers";
import { SnapshotBanner } from "/components/SnapshotBanner";
import type { ComplianceRegistriesPageQuery } from "./__generated__/ComplianceRegistriesPageQuery.graphql";
import type {
  ComplianceRegistriesPageFragment$key,
  ComplianceRegistriesPageFragment$data,
} from "./__generated__/ComplianceRegistriesPageFragment.graphql";

type ComplianceRegistry = ComplianceRegistriesPageFragment$data['complianceRegistries']['edges'][number]['node'];

interface ComplianceRegistriesPageProps {
  queryRef: PreloadedQuery<ComplianceRegistriesPageQuery>;
}

const complianceRegistriesPageFragment = graphql`
  fragment ComplianceRegistriesPageFragment on Organization
  @refetchable(queryName: "ComplianceRegistriesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    id
    complianceRegistries(
      first: $first
      after: $after
      filter: { snapshotId: $snapshotId }
    )
      @connection(key: "ComplianceRegistriesPage_complianceRegistries", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          id
          snapshotId
          sourceId
          referenceId
          area
          source
          requirement
          status
          lastReviewDate
          dueDate
          actionsToBeImplemented
          regulator
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

export default function ComplianceRegistriesPage({ queryRef }: ComplianceRegistriesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("Compliances"));

  const organization = usePreloadedQuery(
    graphql`
      query ComplianceRegistriesPageQuery($organizationId: ID!, $snapshotId: ID) {
        node(id: $organizationId) {
          ... on Organization {
            ...ComplianceRegistriesPageFragment @arguments(snapshotId: $snapshotId)
          }
        }
      }
    `,
    queryRef
  );

  const { data: registriesData, loadNext, hasNext } = usePaginationFragment(
    complianceRegistriesPageFragment,
    organization.node as ComplianceRegistriesPageFragment$key
  );

  const connectionId = registriesData?.complianceRegistries?.__id || "";
  const registries: ComplianceRegistry[] = registriesData?.complianceRegistries?.edges?.map((edge) => edge.node) ?? [];

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <PageHeader
        title={__("Compliances")}
        description={__(
          "Manage your organization's compliance registry entries."
        )}
      >
        {!snapshotId && (
          <CreateComplianceRegistryDialog organizationId={organizationId} connection={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add compliance")}</Button>
          </CreateComplianceRegistryDialog>
        )}
      </PageHeader>

      {registries.length === 0 ? (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No compliance entries yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first compliance entry to get started.")}
            </p>
          </div>
        </Card>
      ) : (
        <Card>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Reference ID")}</Th>
                <Th>{__("Area")}</Th>
                <Th>{__("Source")}</Th>
                <Th>{__("Status")}</Th>
                <Th>{__("Owner")}</Th>
                <Th>{__("Due Date")}</Th>
                <Th>{__("Actions")}</Th>
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
                disabled={!hasNext}
              >
                {__("Load more")}
              </Button>
            </div>
          )}
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
  registry: ComplianceRegistry;
  connectionId: string;
  snapshotId?: string;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const [deleteRegistry] = useMutation(deleteComplianceRegistryMutation);
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
              complianceRegistryId: registry.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the compliance registry entry %s. This action cannot be undone."
          ),
          registry.referenceId
        ),
      }
    );
  };

  const detailsUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/compliance-registries/${registry.id}`
    : `/organizations/${organizationId}/compliance-registries/${registry.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>
        <span className="font-mono text-sm">{registry.referenceId}</span>
      </Td>
      <Td>{registry.area || "-"}</Td>
      <Td>{registry.source || "-"}</Td>
      <Td>
        <Badge variant={getStatusVariant(registry.status || "OPEN")}>
          {getStatusLabel(registry.status || "OPEN")}
        </Badge>
      </Td>
      <Td>{registry.owner?.fullName || "-"}</Td>
      <Td>
        {registry.dueDate ? (
          <time dateTime={registry.dueDate}>
            {formatDate(registry.dueDate)}
          </time>
        ) : (
          <span className="text-txt-tertiary">{__("No due date")}</span>
        )}
      </Td>
      <Td noLink width={50} className="text-end">
        {!isSnapshotMode && (
          <ActionDropdown>
            <DropdownItem
              icon={IconTrashCan}
              variant="danger"
              onSelect={handleDelete}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </Td>
    </Tr>
  );
}
