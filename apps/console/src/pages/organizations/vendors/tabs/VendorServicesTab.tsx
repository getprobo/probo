import { sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { type ComponentProps, useState } from "react";
import { useFragment, useRefetchableFragment } from "react-relay";
import { useOutletContext, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { VendorGraphNodeQuery$data } from "#/__generated__/core/VendorGraphNodeQuery.graphql";
import type { VendorServicesListQuery } from "#/__generated__/core/VendorServicesListQuery.graphql";
import type { VendorServicesTabFragment$key } from "#/__generated__/core/VendorServicesTabFragment.graphql";
import type {
  VendorServicesTabFragment_service$data,
  VendorServicesTabFragment_service$key,
} from "#/__generated__/core/VendorServicesTabFragment_service.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { CreateServiceDialog } from "../dialogs/CreateServiceDialog";
import { EditServiceDialog } from "../dialogs/EditServiceDialog";

export const vendorServicesFragment = graphql`
  fragment VendorServicesTabFragment on Vendor
  @refetchable(queryName: "VendorServicesListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "VendorServiceOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    services(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "VendorServicesTabFragment_services") {
      __id
      edges {
        node {
          id
          canUpdate: permission(action: "core:vendor-service:update")
          canDelete: permission(action: "core:vendor-service:delete")
          ...VendorServicesTabFragment_service
        }
      }
    }
  }
`;

const serviceFragment = graphql`
  fragment VendorServicesTabFragment_service on VendorService {
    id
    name
    description
    createdAt
    updatedAt
    canUpdate: permission(action: "core:vendor-service:update")
    canDelete: permission(action: "core:vendor-service:delete")
  }
`;

const deleteServiceMutation = graphql`
  mutation VendorServicesTabDeleteServiceMutation(
    $input: DeleteVendorServiceInput!
    $connections: [ID!]!
  ) {
    deleteVendorService(input: $input) {
      deletedVendorServiceId @deleteEdge(connections: $connections)
    }
  }
`;

export default function VendorServicesTab() {
  const { vendor } = useOutletContext<{
    vendor: VendorGraphNodeQuery$data["node"];
  }>();
  const [data, refetch] = useRefetchableFragment<
    VendorServicesListQuery,
    VendorServicesTabFragment$key
  >(vendorServicesFragment, vendor);
  const connectionId = data.services.__id;
  const services = data.services.edges.map(edge => edge.node);
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [editingService, setEditingService]
    = useState<VendorServicesTabFragment_service$data | null>(null);
  const hasAnyAction = services.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  usePageTitle(vendor.name + " - " + __("Services"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Services")}
        description={__("Manage services provided by this vendor.")}
      >
        {!isSnapshotMode && vendor.canCreateService && (
          <CreateServiceDialog vendorId={vendor.id} connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add service")}</Button>
          </CreateServiceDialog>
        )}
      </PageHeader>

      <SortableTable
        refetch={refetch as ComponentProps<typeof SortableTable>["refetch"]}
      >
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Name")}</SortableTh>
            <Th>{__("Description")}</Th>
            {!isSnapshotMode && hasAnyAction && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {services.map(service => (
            <ServiceRow
              key={service.id}
              serviceKey={service}
              connectionId={connectionId}
              onEdit={setEditingService}
              isSnapshotMode={isSnapshotMode}
            />
          ))}
        </Tbody>
      </SortableTable>

      {editingService && !isSnapshotMode && editingService.canUpdate && (
        <EditServiceDialog
          serviceId={editingService.id}
          service={editingService}
          onClose={() => setEditingService(null)}
        />
      )}
    </div>
  );
}

type ServiceRowProps = {
  serviceKey: VendorServicesTabFragment_service$key;
  connectionId: string;
  onEdit: (service: VendorServicesTabFragment_service$data) => void;
  isSnapshotMode: boolean;
};

function ServiceRow(props: ServiceRowProps) {
  const { __ } = useTranslate();
  const service = useFragment<VendorServicesTabFragment_service$key>(
    serviceFragment,
    props.serviceKey,
  );
  const confirm = useConfirm();
  const [deleteService] = useMutationWithToasts(deleteServiceMutation, {
    successMessage: __("Service deleted successfully"),
    errorMessage: __("Failed to delete service"),
  });
  const hasAnyAction = service.canUpdate || service.canDelete;

  const handleDelete = () => {
    confirm(
      () =>
        deleteService({
          variables: {
            connections: [props.connectionId],
            input: {
              vendorServiceId: service.id,
            },
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the service \"%s\". This action cannot be undone.",
          ),
          service.name,
        ),
      },
    );
  };

  return (
    <Tr>
      <Td>{service.name}</Td>
      <Td>{service.description || __("—")}</Td>
      {!props.isSnapshotMode && hasAnyAction && (
        <Td width={50} className="text-end">
          <ActionDropdown>
            {service.canUpdate && (
              <DropdownItem
                icon={IconPencil}
                onClick={() => props.onEdit(service)}
              >
                {__("Edit")}
              </DropdownItem>
            )}
            {service.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                onClick={handleDelete}
                variant="danger"
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
