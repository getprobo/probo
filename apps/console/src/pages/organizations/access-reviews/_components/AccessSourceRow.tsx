import { formatDate, formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  DropdownItem,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessSourceRowDeleteMutation } from "#/__generated__/core/AccessSourceRowDeleteMutation.graphql";
import type { AccessSourceRowFragment$key } from "#/__generated__/core/AccessSourceRowFragment.graphql";

const fragment = graphql`
  fragment AccessSourceRowFragment on AccessSource {
    id
    name
    connectorId
    connector {
      provider
    }
    createdAt
    canDelete: permission(action: "core:access-source:delete")
  }
`;

export const deleteAccessSourceMutation = graphql`
  mutation AccessSourceRowDeleteMutation(
    $input: DeleteAccessSourceInput!
    $connections: [ID!]!
  ) {
    deleteAccessSource(input: $input) {
      deletedAccessSourceId @deleteEdge(connections: $connections)
    }
  }
`;

type Props = {
  fKey: AccessSourceRowFragment$key;
  connectionId: string;
};

function sourceLabel(connectorProvider: string | null | undefined): string {
  if (!connectorProvider) {
    return "CSV";
  }

  switch (connectorProvider) {
    case "GOOGLE_WORKSPACE":
      return "Google Workspace";
    case "LINEAR":
      return "Linear";
    case "SLACK":
      return "Slack";
    default:
      return connectorProvider;
  }
}

export function AccessSourceRow({ fKey, connectionId }: Props) {
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const { toast } = useToast();

  const accessSource = useFragment(fragment, fKey);

  const [deleteAccessSource] = useMutation<AccessSourceRowDeleteMutation>(deleteAccessSourceMutation);

  const handleDelete = () => {
    if (!accessSource.id || !accessSource.name) {
      return alert(__("Failed to delete access source: missing id or name"));
    }
    confirm(
      async () => {
        await new Promise<void>((resolve, reject) => {
          deleteAccessSource({
            variables: {
              input: {
                accessSourceId: accessSource.id,
              },
              connections: [connectionId],
            },
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: formatError(
                    __("Failed to delete access source"),
                    errors as GraphQLError[],
                  ),
                  variant: "error",
                });
                reject(new Error(errors[0]?.message ?? __("Failed to delete access source")));
                return;
              }
              resolve();
            },
            onError: (error) => {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete access source"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              reject(error);
            },
          });
        });
      },
      {
        message: sprintf(
          __(
            "This will permanently delete \"%s\". This action cannot be undone.",
          ),
          accessSource.name,
        ),
      },
    );
  };

  return (
    <Tr>
      <Td>{accessSource.name}</Td>
      <Td>
        <Badge variant="neutral" size="sm">
          {sourceLabel(accessSource.connector?.provider ?? null)}
        </Badge>
      </Td>
      <Td>
        <time dateTime={accessSource.createdAt}>
          {formatDate(accessSource.createdAt)}
        </time>
      </Td>
      {accessSource.canDelete && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            <DropdownItem
              icon={IconTrashCan}
              variant="danger"
              onSelect={(e) => {
                e.preventDefault();
                e.stopPropagation();
                handleDelete();
              }}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
