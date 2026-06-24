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
