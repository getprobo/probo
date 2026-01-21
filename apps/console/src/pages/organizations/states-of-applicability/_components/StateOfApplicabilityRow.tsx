import { formatDate, formatError, type GraphQLError, promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { StateOfApplicabilityRowDeleteMutation } from "#/__generated__/core/StateOfApplicabilityRowDeleteMutation.graphql";
import type { StateOfApplicabilityRowFragment$key } from "#/__generated__/core/StateOfApplicabilityRowFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
    fragment StateOfApplicabilityRowFragment on StateOfApplicability {
        id
        name
        createdAt
        canDelete: permission(action: "core:state-of-applicability:delete")
        statementsInfo: applicabilityStatements {
            totalCount
        }
    }
`;

const deleteMutation = graphql`
    mutation StateOfApplicabilityRowDeleteMutation(
        $input: DeleteStateOfApplicabilityInput!
        $connections: [ID!]!
    ) {
        deleteStateOfApplicability(input: $input) {
            deletedStateOfApplicabilityId @deleteEdge(connections: $connections)
        }
    }
`;

type Props = {
  fKey: StateOfApplicabilityRowFragment$key;
  connectionId: string;
};

export function StateOfApplicabilityRow({ fKey, connectionId }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const confirm = useConfirm();
  const { toast } = useToast();
  const isSnapshotMode = Boolean(snapshotId);

  const stateOfApplicability = useFragment(fragment, fKey);
  const canDelete = !isSnapshotMode && stateOfApplicability.canDelete;

  const [deleteStateOfApplicability] = useMutation<StateOfApplicabilityRowDeleteMutation>(deleteMutation);

  const handleDelete = () => {
    if (!stateOfApplicability.id || !stateOfApplicability.name) {
      return alert(__("Failed to delete state of applicability: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(deleteStateOfApplicability)({
          variables: {
            input: {
              stateOfApplicabilityId: stateOfApplicability.id,
            },
            connections: [connectionId],
          },
        }).catch((error) => {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to delete state of applicability"),
              error as GraphQLError,
            ),
            variant: "error",
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete \"%s\". This action cannot be undone.",
          ),
          stateOfApplicability.name,
        ),
      },
    );
  };

  const detailUrl = snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/states-of-applicability/${stateOfApplicability.id}`
    : `/organizations/${organizationId}/states-of-applicability/${stateOfApplicability.id}`;

  return (
    <Tr to={detailUrl}>
      <Td>{stateOfApplicability.name}</Td>
      <Td>
        <time dateTime={stateOfApplicability.createdAt}>
          {formatDate(stateOfApplicability.createdAt)}
        </time>
      </Td>
      <Td>{stateOfApplicability.statementsInfo?.totalCount ?? 0}</Td>
      {canDelete && (
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
