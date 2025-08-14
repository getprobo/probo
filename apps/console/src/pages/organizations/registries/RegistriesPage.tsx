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
  IconPencil,
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
import { useOrganizationId } from "/hooks/useOrganizationId";
import { CreateRegistryDialog } from "./dialogs/CreateRegistryDialog";
import { UpdateRegistryDialog } from "./dialogs/UpdateRegistryDialog";
import { deleteNonconformityRegistryMutation } from "../../../hooks/graph/NonconformityRegistryGraph";
import { useState } from "react";
import { sprintf, promisifyMutation, getStatusVariant, getStatusLabel } from "@probo/helpers";
import type { RegistriesPageQuery } from "./__generated__/RegistriesPageQuery.graphql";
import type {
  RegistriesPageFragment$key,
  RegistriesPageFragment$data,
} from "./__generated__/RegistriesPageFragment.graphql";

type NonconformityRegistry = RegistriesPageFragment$data['nonconformityRegistries']['edges'][number]['node'];

interface RegistriesPageProps {
  queryRef: PreloadedQuery<RegistriesPageQuery>;
}

const registriesPageFragment = graphql`
  fragment RegistriesPageFragment on Organization
  @refetchable(queryName: "RegistriesPageRefetchQuery")
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

export default function RegistriesPage({ queryRef }: RegistriesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const [editingRegistry, setEditingRegistry] = useState<NonconformityRegistry | null>(null);
  const confirm = useConfirm();
  const [deleteRegistry] = useMutation(deleteNonconformityRegistryMutation);

  usePageTitle(__("Registries"));

  const organization = usePreloadedQuery(
    graphql`
      query RegistriesPageQuery($organizationId: ID!) {
        node(id: $organizationId) {
          ... on Organization {
            ...RegistriesPageFragment
          }
        }
      }
    `,
    queryRef
  );

  const { data: registriesData, loadNext, hasNext } = usePaginationFragment(
    registriesPageFragment,
    organization.node as RegistriesPageFragment$key
  );

  const connectionId = registriesData?.nonconformityRegistries.__id;
  const registries: NonconformityRegistry[] = registriesData?.nonconformityRegistries?.edges?.map((edge) => edge.node) ?? [];

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
    <div className="space-y-6">
      <PageHeader
        title={__("Registries")}
        description={__(
          "Manage your organization's nonconformity registries."
        )}
      >
        <CreateRegistryDialog organizationId={organizationId} connection={connectionId}>
          <Button icon={IconPlusLarge}>{__("Add registry entry")}</Button>
        </CreateRegistryDialog>
      </PageHeader>

      {registries.length === 0 ? (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No registry entries yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first nonconformity registry entry to get started.")}
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
                <Th>{__("Date Identified")}</Th>
                <Th>{__("Due Date")}</Th>
                <Th>{__("Root Cause")} *</Th>
                <Th>{__("Corrective Action")}</Th>
                <Th>{__("Effectiveness Check")}</Th>
                <Th>{__("Actions")}</Th>
              </Tr>
            </Thead>
            <Tbody>
              {registries.map((registry) => (
                <Tr key={registry.id}>
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
                  <Td>{registry.audit.framework.name}</Td>
                  <Td>{registry.owner.fullName}</Td>
                  <Td>
                    {registry.dateIdentified ? (
                      <time dateTime={registry.dateIdentified}>
                        {formatDate(registry.dateIdentified)}
                      </time>
                    ) : (
                      <span className="text-txt-tertiary">{__("Not set")}</span>
                    )}
                  </Td>
                  <Td>
                    {registry.dueDate ? (
                      <time dateTime={registry.dueDate}>
                        {formatDate(registry.dueDate)}
                      </time>
                    ) : (
                      <span className="text-txt-tertiary">{__("No due date")}</span>
                    )}
                  </Td>
                  <Td>
                    <div className="min-w-0">
                      <p className="whitespace-pre-wrap break-words">
                        {registry.rootCause || __("No root cause")}
                      </p>
                    </div>
                  </Td>
                  <Td>
                    <div className="min-w-0">
                      <p className="whitespace-pre-wrap break-words">
                        {registry.correctiveAction || __("No corrective action")}
                      </p>
                    </div>
                  </Td>
                  <Td>
                    <div className="min-w-0">
                      <p className="whitespace-pre-wrap break-words">
                        {registry.effectivenessCheck || __("Not checked")}
                      </p>
                    </div>
                  </Td>
                  <Td>
                    <ActionDropdown>
                      <DropdownItem
                        icon={IconPencil}
                        onSelect={() => setEditingRegistry(registry)}
                      >
                        {__("Edit")}
                      </DropdownItem>
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

      {editingRegistry && (
        <UpdateRegistryDialog
          registry={editingRegistry}
          onClose={() => setEditingRegistry(null)}
        />
      )}
    </div>
  );
}
