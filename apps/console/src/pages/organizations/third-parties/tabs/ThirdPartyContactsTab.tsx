// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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

import type { ThirdPartyContactsListQuery } from "#/__generated__/core/ThirdPartyContactsListQuery.graphql";
import type { ThirdPartyContactsTabFragment$key } from "#/__generated__/core/ThirdPartyContactsTabFragment.graphql";
import type {
  ThirdPartyContactsTabFragment_contact$data,
  ThirdPartyContactsTabFragment_contact$key,
} from "#/__generated__/core/ThirdPartyContactsTabFragment_contact.graphql";
import type { ThirdPartyGraphNodeQuery$data } from "#/__generated__/core/ThirdPartyGraphNodeQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { CreateContactDialog } from "../dialogs/CreateContactDialog";
import { EditContactDialog } from "../dialogs/EditContactDialog";

export const thirdPartyContactsFragment = graphql`
  fragment ThirdPartyContactsTabFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyContactsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyContactOrder", defaultValue: null }
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
    ) @connection(key: "ThirdPartyContactsTabFragment_contacts") {
      __id
      edges {
        node {
          id
          canUpdate: permission(action: "core:thirdParty-contact:update")
          canDelete: permission(action: "core:thirdParty-contact:delete")
          ...ThirdPartyContactsTabFragment_contact
        }
      }
    }
  }
`;

const contactFragment = graphql`
  fragment ThirdPartyContactsTabFragment_contact on ThirdPartyContact {
    id
    fullName
    email
    phone
    role
    canUpdate: permission(action: "core:thirdParty-contact:update")
    canDelete: permission(action: "core:thirdParty-contact:delete")
  }
`;

const deleteContactMutation = graphql`
  mutation ThirdPartyContactsTabDeleteContactMutation(
    $input: DeleteThirdPartyContactInput!
    $connections: [ID!]!
  ) {
    deleteThirdPartyContact(input: $input) {
      deletedThirdPartyContactId @deleteEdge(connections: $connections)
    }
  }
`;

export default function ThirdPartyContactsTab() {
  const { thirdParty } = useOutletContext<{
    thirdParty: ThirdPartyGraphNodeQuery$data["node"];
  }>();
  const [data, refetch] = useRefetchableFragment<
    ThirdPartyContactsListQuery,
    ThirdPartyContactsTabFragment$key
  >(thirdPartyContactsFragment, thirdParty);
  const connectionId = data.contacts.__id;
  const contacts = data.contacts.edges.map(edge => edge.node);
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [editingContact, setEditingContact]
    = useState<ThirdPartyContactsTabFragment_contact$data | null>(null);
  const hasAnyAction = contacts.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  usePageTitle(thirdParty.name + " - " + __("Contacts"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Contacts")}
        description={__("Manage thirdParty contacts and their information.")}
      >
        {!isSnapshotMode && thirdParty.canCreateContact && (
          <CreateContactDialog thirdPartyId={thirdParty.id} connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add contact")}</Button>
          </CreateContactDialog>
        )}
      </PageHeader>

      <SortableTable
        refetch={refetch as ComponentProps<typeof SortableTable>["refetch"]}
      >
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
          {contacts.map(contact => (
            <ContactRow
              key={contact.id}
              contactKey={contact}
              connectionId={connectionId}
              onEdit={setEditingContact}
              isSnapshotMode={isSnapshotMode}
            />
          ))}
        </Tbody>
      </SortableTable>

      {editingContact && !isSnapshotMode && editingContact.canUpdate && (
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
  contactKey: ThirdPartyContactsTabFragment_contact$key;
  connectionId: string;
  onEdit: (contact: ThirdPartyContactsTabFragment_contact$data) => void;
  isSnapshotMode: boolean;
};

function ContactRow(props: ContactRowProps) {
  const { __ } = useTranslate();
  const contact = useFragment<ThirdPartyContactsTabFragment_contact$key>(
    contactFragment,
    props.contactKey,
  );
  const confirm = useConfirm();
  const [deleteContact] = useMutationWithToasts(deleteContactMutation, {
    successMessage: __("Contact deleted successfully"),
    errorMessage: __("Failed to delete contact"),
  });
  const hasAnyAction = contact.canUpdate || contact.canDelete;

  const handleDelete = () => {
    confirm(
      () =>
        deleteContact({
          variables: {
            connections: [props.connectionId],
            input: {
              thirdPartyContactId: contact.id,
            },
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the contact \"%s\". This action cannot be undone.",
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
        {contact.email
          ? (
              <a
                href={`mailto:${contact.email}`}
                className="text-primary-600 hover:text-primary-800"
              >
                {contact.email}
              </a>
            )
          : (
              __("—")
            )}
      </Td>
      <Td>
        {contact.phone
          ? (
              <a
                href={`tel:${contact.phone}`}
                className="text-primary-600 hover:text-primary-800"
              >
                {contact.phone}
              </a>
            )
          : (
              __("—")
            )}
      </Td>
      <Td>{contact.role || __("—")}</Td>
      {!props.isSnapshotMode && hasAnyAction && (
        <Td width={50} className="text-end">
          <ActionDropdown>
            {contact.canUpdate && (
              <DropdownItem
                icon={IconPencil}
                onClick={() => props.onEdit(contact)}
              >
                {__("Edit")}
              </DropdownItem>
            )}
            {contact.canDelete && (
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
