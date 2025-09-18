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
  ConnectionHandler,
  type PreloadedQuery,
} from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { CreateNonconformityDialog } from "./dialogs/CreateNonconformityDialog";
import { deleteNonconformityMutation, NonconformitiesConnectionKey } from "../../../hooks/graph/NonconformityGraph";
import { sprintf, promisifyMutation, getStatusVariant, getStatusLabel, formatDate } from "@probo/helpers";
import { SnapshotBanner } from "/components/SnapshotBanner";
import { useParams } from "react-router";
import type { NonconformitiesPageQuery } from "./__generated__/NonconformitiesPageQuery.graphql";
import type {
  NonconformitiesPageFragment$key,
  NonconformitiesPageFragment$data,
} from "./__generated__/NonconformitiesPageFragment.graphql";

type Nonconformity = NonconformitiesPageFragment$data['nonconformities']['edges'][number]['node'];

interface NonconformitiesPageProps {
  queryRef: PreloadedQuery<NonconformitiesPageQuery>;
}

const nonconformitiesPageFragment = graphql`
  fragment NonconformitiesPageFragment on Organization
  @refetchable(queryName: "NonconformitiesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    id
    nonconformities(
      first: $first
      after: $after
      filter: { snapshotId: $snapshotId }
    )
      @connection(key: "NonconformitiesPage_nonconformities", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          id
          referenceId
          snapshotId
          description
          status
          dateIdentified
          dueDate
          rootCause
          correctiveAction
          effectivenessCheck
          audit {
            id
            name
            framework {
              id
              name
            }
          }
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

export default function NonconformitiesPage({ queryRef }: NonconformitiesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("Nonconformities"));

  const organization = usePreloadedQuery(
    graphql`
      query NonconformitiesPageQuery($organizationId: ID!, $snapshotId: ID) {
        node(id: $organizationId) {
          ... on Organization {
            ...NonconformitiesPageFragment @arguments(snapshotId: $snapshotId)
          }
        }
      }
    `,
    queryRef
  );

  const { data: nonconformitiesData, loadNext, hasNext } = usePaginationFragment(
    nonconformitiesPageFragment,
    organization.node as NonconformitiesPageFragment$key
  );

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    NonconformitiesConnectionKey,
    { filter: { snapshotId: snapshotId || null } }
  );
  const nonconformities: Nonconformity[] = nonconformitiesData?.nonconformities?.edges?.map((edge) => edge.node) ?? [];

  return (
    <div className="space-y-6">
      {isSnapshotMode && (
        <SnapshotBanner snapshotId={snapshotId!} />
      )}
      <PageHeader
        title={__("Nonconformities")}
        description={__(
          "Manage your organization's non conformities."
        )}
      >
        {!isSnapshotMode && (
          <CreateNonconformityDialog organizationId={organizationId} connection={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add nonconformity")}</Button>
          </CreateNonconformityDialog>
        )}
      </PageHeader>

      {nonconformities.length === 0 ? (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No nonconformities yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first nonconformity to get started.")}
            </p>
          </div>
        </Card>
      ) : (
        <Card>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Reference ID")}</Th>
                <Th>{__("Description")}</Th>
                <Th>{__("Status")}</Th>
                <Th>{__("Audit")}</Th>
                <Th>{__("Owner")}</Th>
                <Th>{__("Due Date")}</Th>
                {!isSnapshotMode && (<Th>{__("Actions")}</Th>)}
              </Tr>
            </Thead>
            <Tbody>
              {nonconformities.map((nonconformity) => (
                <NonconformityRow
                  key={nonconformity.id}
                  nonconformity={nonconformity}
                  connectionId={connectionId}
                  isSnapshotMode={isSnapshotMode}
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

function NonconformityRow({
  nonconformity,
  connectionId,
  isSnapshotMode,
  snapshotId,
}: {
  nonconformity: Nonconformity;
  connectionId: string;
  isSnapshotMode: boolean;
  snapshotId?: string;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [deleteNonconformity] = useMutation(deleteNonconformityMutation);


  const nonconformityDetailUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/nonconformities/${nonconformity.id}`
    : `/organizations/${organizationId}/nonconformities/${nonconformity.id}`;

  const handleDeleteNonconformity = (nonconformity: Nonconformity) => {
    if (!connectionId) return;

    confirm(
      () => {
        return promisifyMutation(deleteNonconformity)({
          variables: {
            input: {
              nonconformityId: nonconformity.id,
            },
            connections: [connectionId],
          },
        });
      },
      {
        message: sprintf(
          __(
            "This will permanently delete the nonconformity %s. This action cannot be undone."
          ),
          nonconformity.referenceId
        ),
      }
    );
  };

  return (
    <Tr to={nonconformityDetailUrl}>
      <Td>
        <span className="font-mono text-sm">{nonconformity.referenceId}</span>
      </Td>
      <Td>
        <div className="min-w-0">
          <p className="whitespace-pre-wrap break-words">
            {nonconformity.description || __("No description")}
          </p>
        </div>
      </Td>
      <Td>
        <Badge variant={getStatusVariant(nonconformity.status)}>
          {getStatusLabel(nonconformity.status)}
        </Badge>
      </Td>
      <Td>
        {nonconformity.audit.name
          ? `${nonconformity.audit.framework.name} - ${nonconformity.audit.name}`
          : nonconformity.audit.framework.name
        }
      </Td>
      <Td>{nonconformity.owner.fullName}</Td>
      <Td>
        {nonconformity.dueDate ? (
          <time dateTime={nonconformity.dueDate}>
            {formatDate(nonconformity.dueDate)}
          </time>
        ) : (
          <span className="text-txt-tertiary">{__("No due date")}</span>
        )}
      </Td>
      {!isSnapshotMode && (<Td noLink width={50} className="text-end">
          <ActionDropdown>
            <DropdownItem
              icon={IconTrashCan}
              variant="danger"
              onSelect={() => handleDeleteNonconformity(nonconformity)}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
