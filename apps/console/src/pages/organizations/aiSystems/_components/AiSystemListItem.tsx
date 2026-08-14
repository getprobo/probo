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

import { dateFormat } from "@probo/i18n";
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
import { useTranslation } from "react-i18next";
import { ConnectionHandler, graphql, useFragment } from "react-relay";

import type { AiSystemListItem_aiSystem$key } from "#/__generated__/core/AiSystemListItem_aiSystem.graphql";
import type { AiSystemListItemDeleteMutation } from "#/__generated__/core/AiSystemListItemDeleteMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  aiSystemListConnectionFilters,
  AiSystemsConnectionKey,
  getRiskClassificationLabel,
  getRiskClassificationVariant,
  getStatusLabel,
  getStatusVariant,
} from "../_lib/aiSystemHelpers";

const aiSystemListItemFragment = graphql`
  fragment AiSystemListItem_aiSystem on AiSystem {
    id
    name
    version
    status
    riskClassification
    nextReviewDate
    owner {
      id
      fullName
    }
    canDelete: permission(action: "core:ai-system:delete")
  }
`;

const deleteAiSystemMutation = graphql`
  mutation AiSystemListItemDeleteMutation(
    $input: DeleteAiSystemInput!
    $connections: [ID!]!
  ) {
    deleteAiSystem(input: $input) {
      deletedAiSystemId @deleteEdge(connections: $connections)
    }
  }
`;

interface AiSystemListItemProps {
  aiSystemKey: AiSystemListItem_aiSystem$key;
  hasAnyAction: boolean;
}

export function AiSystemListItem({
  aiSystemKey,
  hasAnyAction,
}: AiSystemListItemProps) {
  const aiSystem = useFragment(aiSystemListItemFragment, aiSystemKey);
  const organizationId = useOrganizationId();
  const { t, i18n } = useTranslation();
  const confirm = useConfirm();
  const { toast } = useToast();
  const [deleteAiSystem] = useMutation<AiSystemListItemDeleteMutation>(
    deleteAiSystemMutation,
    {
      errorToast: t("aiSystemListItem.errors.delete"),
    },
  );

  const deleteConnections = aiSystemListConnectionFilters(aiSystem).map(filter =>
    ConnectionHandler.getConnectionID(
      organizationId,
      AiSystemsConnectionKey,
      { filter },
    ),
  );

  const handleDelete = () => {
    confirm(
      async () => {
        await deleteAiSystem({
          variables: {
            input: {
              aiSystemId: aiSystem.id,
            },
            connections: deleteConnections,
          },
        });
        toast({
          title: t("aiSystemListItem.messages.success"),
          description: t("aiSystemListItem.messages.deleted"),
          variant: "success",
        });
      },
      {
        message: t("aiSystemListItem.deleteConfirmation", {
          name: aiSystem.name,
        }),
      },
    );
  };

  const detailsUrl = `/organizations/${organizationId}/registries/ai-systems/${aiSystem.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>{aiSystem.name}</Td>
      <Td>{aiSystem.version || "-"}</Td>
      <Td>
        <Badge variant={getStatusVariant(aiSystem.status)}>
          {getStatusLabel(aiSystem.status, t, "aiSystemsPage")}
        </Badge>
      </Td>
      <Td>
        {aiSystem.riskClassification
          ? (
              <Badge variant={getRiskClassificationVariant(aiSystem.riskClassification)}>
                {getRiskClassificationLabel(
                  aiSystem.riskClassification,
                  t,
                  "aiSystemsPage",
                )}
              </Badge>
            )
          : (
              <span className="text-txt-secondary">-</span>
            )}
      </Td>
      <Td>{aiSystem.owner?.fullName || "-"}</Td>
      <Td>
        {aiSystem.nextReviewDate
          ? (
              <time dateTime={aiSystem.nextReviewDate.split("T")[0]}>
                {dateFormat(i18n.language, aiSystem.nextReviewDate.split("T")[0], {
                  year: "numeric",
                  month: "short",
                  day: "numeric",
                })}
              </time>
            )
          : (
              <span className="text-txt-secondary">-</span>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          {aiSystem.canDelete && (
            <ActionDropdown>
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {t("aiSystemListItem.actions.delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </Td>
      )}
    </Tr>
  );
}
