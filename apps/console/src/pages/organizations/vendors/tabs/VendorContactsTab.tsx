import { useOutletContext, useParams } from "react-router";
import { graphql } from "relay-runtime";
import type { VendorContactsTabFragment$key } from "/__generated__/core/VendorContactsTabFragment.graphql";
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
import type { VendorContactsTabFragment_contact$key } from "/__generated__/core/VendorContactsTabFragment_contact.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { sprintf } from "@probo/helpers";
import { SortableTable, SortableTh } from "/components/SortableTable";
import { CreateContactDialog } from "../dialogs/CreateContactDialog";
import { EditContactDialog } from "../dialogs/EditContactDialog";
import { use, useState } from "react";
import { PermissionsContext } from "/providers/PermissionsContext";

export const vendorContactsFragment = graphql`
  fragment VendorContactsTabFragment on Vendor
  @refetchable(queryName: "VendorContactsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "VendorContactOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    contacts(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "VendorContactsTabFragment_contacts") {
      __id
      edges {
        node {
          id
          ...VendorContactsTabFragment_contact
        }
      }
    }
  }
`;

const contactFragment = graphql`
  fragment VendorContactsTabFragment_contact on VendorContact {
    id
    fullName
    email
    phone
    role
    createdAt
    updatedAt
  }
`;

const deleteContactMutation = graphql`
  mutation VendorContactsTabDeleteContactMutation(
    $input: DeleteVendorContactInput!
    $connections: [ID!]!
  ) {
    deleteVendorContact(input: $input) {
      deletedVendorContactId @deleteEdge(connections: $connections)
    }
  }
`;

export default function VendorContactsTab() {
  const { vendor } = useOutletContext<{
    vendor: VendorContactsTabFragment$key & { name: string; id: string };
  }>();
  const [data, refetch] = useRefetchableFragment(
    vendorContactsFragment,
    vendor,
  );
  const connectionId = data.contacts.__id;
  const contacts = data.contacts.edges.map((edge) => edge.node);
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [editingContact, setEditingContact] = useState<{
    id: string;
    fullName?: string | null;
    email?: string | null;
    phone?: string | null;
    role?: string | null;
  } | null>(null);
  const { isAuthorized } = use(PermissionsContext);
  const canCreateContact = isAuthorized("Vendor", "createVendorContact");
  const canUpdateContact = isAuthorized("VendorContact", "updateVendorContact");
  const canDeleteContact = isAuthorized("VendorContact", "deleteVendorContact");
  const hasAnyAction = canUpdateContact || canDeleteContact;

  usePageTitle(vendor.name + " - " + __("Contacts"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Contacts")}
        description={__("Manage vendor contacts and their information.")}
      >
        {!isSnapshotMode && canCreateContact && (
          <CreateContactDialog vendorId={vendor.id} connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add contact")}</Button>
          </CreateContactDialog>
        )}
      </PageHeader>

      <SortableTable refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
            <SortableTh field="EMAIL">{__("Email")}</SortableTh>
            <Th>{__("Phone")}</Th>
            <Th>{__("Role")}</Th>
            {!isSnapshotMode && hasAnyAction && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {contacts.map((contact) => (
            <ContactRow
              key={contact.id}
              contactKey={contact}
              connectionId={connectionId}
              onEdit={setEditingContact}
              isSnapshotMode={isSnapshotMode}
              canUpdate={canUpdateContact}
              canDelete={canDeleteContact}
            />
          ))}
        </Tbody>
      </SortableTable>

      {editingContact && !isSnapshotMode && canUpdateContact && (
        <EditContactDialog
          contactId={editingContact.id}
          contact={editingContact}
          onClose={() => setEditingContact(null)}
        />
      )}
    </div>
  );
}

type ContactRowProps = {
  contactKey: VendorContactsTabFragment_contact$key;
  connectionId: string;
  onEdit: (contact: {
    id: string;
    fullName?: string | null;
    email?: string | null;
    phone?: string | null;
    role?: string | null;
  }) => void;
  isSnapshotMode: boolean;
  canUpdate?: boolean;
  canDelete?: boolean;
};

function ContactRow(props: ContactRowProps) {
  const { __ } = useTranslate();
  const contact = useFragment<VendorContactsTabFragment_contact$key>(
    contactFragment,
    props.contactKey,
  );
  const confirm = useConfirm();
  const [deleteContact] = useMutationWithToasts(deleteContactMutation, {
    successMessage: __("Contact deleted successfully"),
    errorMessage: __("Failed to delete contact"),
  });
  const hasAnyAction = props.canUpdate || props.canDelete;

  const handleDelete = () => {
    confirm(
      () =>
        deleteContact({
          variables: {
            connections: [props.connectionId],
            input: {
              vendorContactId: contact.id,
            },
          },
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the contact "%s". This action cannot be undone.',
          ),
          contact.fullName || contact.email || __("Unnamed contact"),
        ),
      },
    );
  };

  return (
    <Tr>
      <Td>{contact.fullName || __("—")}</Td>
      <Td>
        {contact.email ? (
          <a
            href={`mailto:${contact.email}`}
            className="text-primary-600 hover:text-primary-800"
          >
            {contact.email}
          </a>
        ) : (
          __("—")
        )}
      </Td>
      <Td>
        {contact.phone ? (
          <a
            href={`tel:${contact.phone}`}
            className="text-primary-600 hover:text-primary-800"
          >
            {contact.phone}
          </a>
        ) : (
          __("—")
        )}
      </Td>
      <Td>{contact.role || __("—")}</Td>
      {!props.isSnapshotMode && hasAnyAction && (
        <Td width={50} className="text-end">
          <ActionDropdown>
            {props.canUpdate && (
              <DropdownItem
                icon={IconPencil}
                onClick={() =>
                  props.onEdit({
                    id: contact.id,
                    fullName: contact.fullName,
                    email: contact.email,
                    phone: contact.phone,
                    role: contact.role,
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
