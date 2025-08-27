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
import { CreateNonconformityRegistryDialog } from "./dialogs/CreateNonconformityRegistryDialog";
import { deleteNonconformityRegistryMutation, RegistriesConnectionKey } from "../../../hooks/graph/NonconformityRegistryGraph";
import { sprintf, promisifyMutation, getStatusVariant, getStatusLabel } from "@probo/helpers";
import type { NonconformityRegistriesPageQuery } from "./__generated__/NonconformityRegistriesPageQuery.graphql";
import type {
  NonconformityRegistriesPageFragment$key,
  NonconformityRegistriesPageFragment$data,
} from "./__generated__/NonconformityRegistriesPageFragment.graphql";

type NonconformityRegistry = NonconformityRegistriesPageFragment$data['nonconformityRegistries']['edges'][number]['node'];

interface NonconformityRegistriesPageProps {
  queryRef: PreloadedQuery<NonconformityRegistriesPageQuery>;
}

const nonconformityRegistriesPageFragment = graphql`
  fragment NonconformityRegistriesPageFragment on Organization
  @refetchable(queryName: "NonconformityRegistriesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
  ) {
    id
    nonconformityRegistries(first: $first, after: $after)
      @connection(key: "RegistriesPage_nonconformityRegistries") {
      __id
      totalCount
      edges {
        node {
          id
          referenceId
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

export default function NonconformityRegistriesPage({ queryRef }: NonconformityRegistriesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  usePageTitle(__("Nonconformity Registries"));

  const organization = usePreloadedQuery(
    graphql`
      query NonconformityRegistriesPageQuery($organizationId: ID!) {
        node(id: $organizationId) {
          ... on Organization {
            ...NonconformityRegistriesPageFragment
          }
        }
      }
    `,
    queryRef
  );

  const { data: registriesData, loadNext, hasNext } = usePaginationFragment(
    nonconformityRegistriesPageFragment,
    organization.node as NonconformityRegistriesPageFragment$key
  );

  const connectionId = ConnectionHandler.getConnectionID(organizationId, RegistriesConnectionKey);
  const registries: NonconformityRegistry[] = registriesData?.nonconformityRegistries?.edges?.map((edge) => edge.node) ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Nonconformity Registries")}
        description={__(
          "Manage your organization's non conformity registries."
        )}
      >
        <CreateNonconformityRegistryDialog organizationId={organizationId} connection={connectionId}>
          <Button icon={IconPlusLarge}>{__("Add nonconformity registry")}</Button>
        </CreateNonconformityRegistryDialog>
      </PageHeader>

      {registries.length === 0 ? (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No non conformity registry entries yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first non conformity registry entry to get started.")}
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
                <Th>{__("Actions")}</Th>
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
}: {
  registry: NonconformityRegistry;
  connectionId: string;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [deleteRegistry] = useMutation(deleteNonconformityRegistryMutation);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const handleDeleteRegistry = (registry: NonconformityRegistry) => {
    if (!connectionId) return;

    confirm(
      () => {
        return promisifyMutation(deleteRegistry)({
          variables: {
            input: {
              nonconformityRegistryId: registry.id,
            },
            connections: [connectionId],
          },
        });
      },
      {
        message: sprintf(
          __(
            "This will permanently delete the registry entry %s. This action cannot be undone."
          ),
          registry.referenceId
        ),
      }
    );
  };

  return (
    <Tr to={`/organizations/${organizationId}/nonconformity-registries/${registry.id}`}>
      <Td>
        <span className="font-mono text-sm">{registry.referenceId}</span>
      </Td>
      <Td>
        <div className="min-w-0">
          <p className="whitespace-pre-wrap break-words">
            {registry.description || __("No description")}
          </p>
        </div>
      </Td>
      <Td>
        <Badge variant={getStatusVariant(registry.status)}>
          {getStatusLabel(registry.status)}
        </Badge>
      </Td>
      <Td>
        {registry.audit.name
          ? `${registry.audit.framework.name} - ${registry.audit.name}`
          : registry.audit.framework.name
        }
      </Td>
      <Td>{registry.owner.fullName}</Td>
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
        <ActionDropdown>
          <DropdownItem
            icon={IconTrashCan}
            variant="danger"
            onSelect={() => handleDeleteRegistry(registry)}
          >
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </Td>
    </Tr>
  );
}
