import { useOutletContext, useParams } from "react-router";
import { graphql } from "relay-runtime";
import type { VendorServicesTabFragment$key } from "/__generated__/core/VendorServicesTabFragment.graphql";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  IconPencil,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { useFragment, useRefetchableFragment } from "react-relay";
import type { VendorServicesTabFragment_service$key } from "/__generated__/core/VendorServicesTabFragment_service.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { sprintf } from "@probo/helpers";
import { SortableTable, SortableTh } from "/components/SortableTable";
import { CreateServiceDialog } from "../dialogs/CreateServiceDialog";
import { EditServiceDialog } from "../dialogs/EditServiceDialog";
import { use, useState } from "react";
import { PermissionsContext } from "/providers/PermissionsContext";

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
    vendor: VendorServicesTabFragment$key & { name: string; id: string };
  }>();
  const [data, refetch] = useRefetchableFragment(
    vendorServicesFragment,
    vendor,
  );
  const connectionId = data.services.__id;
  const services = data.services.edges.map((edge) => edge.node);
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [editingService, setEditingService] = useState<{
    id: string;
    name: string;
    description?: string | null;
  } | null>(null);
  const { isAuthorized } = use(PermissionsContext);
  const canCreateService = isAuthorized("Vendor", "createVendorService");
  const canUpdateService = isAuthorized("VendorService", "updateVendorService");
  const canDeleteService = isAuthorized("VendorService", "deleteVendorService");
  const hasAnyAction = canUpdateService || canDeleteService;

  usePageTitle(vendor.name + " - " + __("Services"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Services")}
        description={__("Manage services provided by this vendor.")}
      >
        {!isSnapshotMode && canCreateService && (
          <CreateServiceDialog vendorId={vendor.id} connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add service")}</Button>
          </CreateServiceDialog>
        )}
      </PageHeader>

      <SortableTable refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Name")}</SortableTh>
            <Th>{__("Description")}</Th>
            {!isSnapshotMode && hasAnyAction && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {services.map((service) => (
            <ServiceRow
              key={service.id}
              serviceKey={service}
              connectionId={connectionId}
              onEdit={setEditingService}
              isSnapshotMode={isSnapshotMode}
              canUpdate={canUpdateService}
              canDelete={canDeleteService}
            />
          ))}
        </Tbody>
      </SortableTable>

      {editingService && !isSnapshotMode && canUpdateService && (
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
  onEdit: (service: {
    id: string;
    name: string;
    description?: string | null;
  }) => void;
  isSnapshotMode: boolean;
  canUpdate?: boolean;
  canDelete?: boolean;
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
  const hasAnyAction = props.canUpdate || props.canDelete;

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
            'This will permanently delete the service "%s". This action cannot be undone.',
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
            {props.canUpdate && (
              <DropdownItem
                icon={IconPencil}
                onClick={() =>
                  props.onEdit({
                    id: service.id,
                    name: service.name,
                    description: service.description,
                  })
                }
              >
                {__("Edit")}
              </DropdownItem>
            )}
            {props.canDelete && (
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
