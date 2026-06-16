// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { graphql, useFragment, useMutation } from "react-relay";

import type { ThirdPartyContactRowDeleteMutation } from "#/__generated__/core/ThirdPartyContactRowDeleteMutation.graphql";
import type {
  ThirdPartyContactRow_contact$data,
  ThirdPartyContactRow_contact$key,
} from "#/__generated__/core/ThirdPartyContactRow_contact.graphql";

const contactRowFragment = graphql`
  fragment ThirdPartyContactRow_contact on ThirdPartyContact {
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
  mutation ThirdPartyContactRowDeleteMutation(
    $input: DeleteThirdPartyContactInput!
    $connections: [ID!]!
  ) {
    deleteThirdPartyContact(input: $input) {
      deletedThirdPartyContactId @deleteEdge(connections: $connections)
    }
  }
`;

interface ThirdPartyContactRowProps {
  contactKey: ThirdPartyContactRow_contact$key;
  connectionId: string;
  onEdit: (contact: ThirdPartyContactRow_contact$data) => void;
}

export function ThirdPartyContactRow(props: ThirdPartyContactRowProps) {
  const { __ } = useTranslate();
  const contact = useFragment(contactRowFragment, props.contactKey);
  const confirm = useConfirm();
  const { toast } = useToast();
  const [deleteContact] = useMutation<ThirdPartyContactRowDeleteMutation>(
    deleteContactMutation,
  );
  const hasAnyAction = contact.canUpdate || contact.canDelete;

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteContact({
            variables: {
              connections: [props.connectionId],
              input: { thirdPartyContactId: contact.id },
            },
            onCompleted() {
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete contact"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
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
      {hasAnyAction && (
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
