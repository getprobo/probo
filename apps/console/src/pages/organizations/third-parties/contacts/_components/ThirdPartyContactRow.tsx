// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { formatError } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { ThirdPartyContactRow_contact$key } from "#/__generated__/core/ThirdPartyContactRow_contact.graphql";
import type { ThirdPartyContactRowDeleteMutation } from "#/__generated__/core/ThirdPartyContactRowDeleteMutation.graphql";

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
  onEdit: () => void;
}

export function ThirdPartyContactRow(props: ThirdPartyContactRowProps) {
  const { t } = useTranslation();
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
                title: t("thirdPartyContactRow.messages.error"),
                description: formatError(
                  t("thirdPartyContactRow.errors.delete"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("thirdPartyContactRow.deleteConfirmation", {
          name: contact.fullName || contact.email || t("thirdPartyContactRow.unnamed"),
        }),
      },
    );
  };

  return (
    <Tr>
      <Td>{contact.fullName || t("thirdPartyContactRow.emptyValue")}</Td>
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
              t("thirdPartyContactRow.emptyValue")
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
              t("thirdPartyContactRow.emptyValue")
            )}
      </Td>
      <Td>{contact.role || t("thirdPartyContactRow.emptyValue")}</Td>
      {hasAnyAction && (
        <Td width={50} className="text-end">
          <ActionDropdown>
            {contact.canUpdate && (
              <DropdownItem
                icon={IconPencil}
                onClick={() => props.onEdit()}
              >
                {t("thirdPartyContactRow.actions.edit")}
              </DropdownItem>
            )}
            {contact.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                onClick={handleDelete}
                variant="danger"
              >
                {t("thirdPartyContactRow.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
