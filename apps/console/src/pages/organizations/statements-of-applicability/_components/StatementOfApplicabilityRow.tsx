// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { formatError, type GraphQLError, promisifyMutation } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
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
  const { t, i18n } = useTranslation();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const { toast } = useToast();

  const statementOfApplicability = useFragment(fragment, fKey);
  const canDelete = statementOfApplicability.canDelete;

  const [deleteStatementOfApplicability] = useMutation<StatementOfApplicabilityRowDeleteMutation>(deleteMutation);

  const handleDelete = () => {
    if (!statementOfApplicability.id || !statementOfApplicability.name) {
      return alert(t("statementOfApplicabilityDetailPage.errors.missingDeleteData"));
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
            title: t("statementOfApplicabilityDetailPage.messages.error"),
            description: formatError(
              t("statementOfApplicabilityDetailPage.errors.delete"),
              error as GraphQLError,
            ),
            variant: "error",
          });
        }),
      {
        message: t("statementOfApplicabilityDetailPage.deleteConfirmation", { name: statementOfApplicability.name }),
      },
    );
  };

  const detailUrl = `/organizations/${organizationId}/statements-of-applicability/${statementOfApplicability.id}`;

  return (
    <Tr to={detailUrl}>
      <Td>{statementOfApplicability.name}</Td>
      <Td>
        <time dateTime={statementOfApplicability.createdAt}>
          {dateFormat(i18n.language, statementOfApplicability.createdAt)}
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
              {t("statementOfApplicabilityDetailPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
