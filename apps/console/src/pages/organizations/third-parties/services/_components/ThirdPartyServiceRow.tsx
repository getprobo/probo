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

import { formatError, sprintf } from "@probo/helpers";
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

import type { ThirdPartyServiceRow_service$key } from "#/__generated__/core/ThirdPartyServiceRow_service.graphql";
import type { ThirdPartyServiceRowDeleteMutation } from "#/__generated__/core/ThirdPartyServiceRowDeleteMutation.graphql";

const serviceRowFragment = graphql`
  fragment ThirdPartyServiceRow_service on ThirdPartyService {
    id
    name
    description
    canUpdate: permission(action: "core:thirdParty-service:update")
    canDelete: permission(action: "core:thirdParty-service:delete")
  }
`;

const deleteServiceMutation = graphql`
  mutation ThirdPartyServiceRowDeleteMutation(
    $input: DeleteThirdPartyServiceInput!
    $connections: [ID!]!
  ) {
    deleteThirdPartyService(input: $input) {
      deletedThirdPartyServiceId @deleteEdge(connections: $connections)
    }
  }
`;

interface ThirdPartyServiceRowProps {
  serviceKey: ThirdPartyServiceRow_service$key;
  connectionId: string;
  onEdit: () => void;
}

export function ThirdPartyServiceRow(props: ThirdPartyServiceRowProps) {
  const { __ } = useTranslate();
  const service = useFragment(serviceRowFragment, props.serviceKey);
  const confirm = useConfirm();
  const { toast } = useToast();
  const [deleteService] = useMutation<ThirdPartyServiceRowDeleteMutation>(
    deleteServiceMutation,
  );
  const hasAnyAction = service.canUpdate || service.canDelete;

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteService({
            variables: {
              connections: [props.connectionId],
              input: { thirdPartyServiceId: service.id },
            },
            onCompleted(_response, errors) {
              if (errors) {
                toast({
                  title: __("Error"),
                  description: formatError(
                    __("Failed to delete service"),
                    errors,
                  ),
                  variant: "error",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete service"),
                  error,
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
      {hasAnyAction && (
        <Td width={50} className="text-end">
          <ActionDropdown>
            {service.canUpdate && (
              <DropdownItem
                icon={IconPencil}
                onClick={() => props.onEdit()}
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
