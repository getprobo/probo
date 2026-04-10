// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
import { graphql } from "relay-runtime";

import type { StatementOfApplicabilityRowDeleteMutation } from "#/__generated__/core/StatementOfApplicabilityRowDeleteMutation.graphql";
import type { StatementOfApplicabilityRowFragment$key } from "#/__generated__/core/StatementOfApplicabilityRowFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
    fragment StatementOfApplicabilityRowFragment on StatementOfApplicability {
        id
        name
        createdAt
        canDelete: permission(action: "core:statement-of-applicability:delete")
        statementsInfo: applicabilityStatements {
            totalCount
        }
    }
`;

const deleteMutation = graphql`
    mutation StatementOfApplicabilityRowDeleteMutation(
        $input: DeleteStatementOfApplicabilityInput!
        $connections: [ID!]!
    ) {
        deleteStatementOfApplicability(input: $input) {
            deletedStatementOfApplicabilityId @deleteEdge(connections: $connections)
        }
    }
`;

type Props = {
  fKey: StatementOfApplicabilityRowFragment$key;
  connectionId: string;
};

export function StatementOfApplicabilityRow({ fKey, connectionId }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const { toast } = useToast();

  const statementOfApplicability = useFragment(fragment, fKey);
  const canDelete = statementOfApplicability.canDelete;

  const [deleteStatementOfApplicability] = useMutation<StatementOfApplicabilityRowDeleteMutation>(deleteMutation);

  const handleDelete = () => {
    if (!statementOfApplicability.id || !statementOfApplicability.name) {
      return alert(__("Failed to delete statement of applicability: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(deleteStatementOfApplicability)({
          variables: {
            input: {
              statementOfApplicabilityId: statementOfApplicability.id,
            },
            connections: [connectionId],
          },
        }).catch((error) => {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to delete statement of applicability"),
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
          statementOfApplicability.name,
        ),
      },
    );
  };

  const detailUrl = `/organizations/${organizationId}/statements-of-applicability/${statementOfApplicability.id}`;

  return (
    <Tr to={detailUrl}>
      <Td>{statementOfApplicability.name}</Td>
      <Td>
        <time dateTime={statementOfApplicability.createdAt}>
          {formatDate(statementOfApplicability.createdAt)}
        </time>
      </Td>
      <Td>{statementOfApplicability.statementsInfo?.totalCount ?? 0}</Td>
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
