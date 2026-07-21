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

import { faviconUrl, formatError } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  DropdownItem,
  IconTrashCan,
  RiskBadge,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { ThirdPartyRow_thirdParty$key } from "#/__generated__/core/ThirdPartyRow_thirdParty.graphql";
import type { ThirdPartyRowDeleteMutation } from "#/__generated__/core/ThirdPartyRowDeleteMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const thirdPartyRowFragment = graphql`
  fragment ThirdPartyRow_thirdParty on ThirdParty {
    id
    name
    websiteUrl
    riskAssessments(
      first: 1
      orderBy: { direction: DESC, field: CREATED_AT }
    ) {
      edges {
        node {
          createdAt
          dataSensitivity
          businessImpact
        }
      }
    }
    canDelete: permission(action: "core:thirdParty:delete")
  }
`;

const deleteThirdPartyMutation = graphql`
  mutation ThirdPartyRowDeleteMutation(
    $input: DeleteThirdPartyInput!
    $connections: [ID!]!
  ) {
    deleteThirdParty(input: $input) {
      deletedThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

interface ThirdPartyRowProps {
  thirdPartyKey: ThirdPartyRow_thirdParty$key;
  connectionId: string;
  hasAnyAction: boolean;
}

export function ThirdPartyRow(props: ThirdPartyRowProps) {
  const { t, i18n } = useTranslation();
  const organizationId = useOrganizationId();
  const thirdParty = useFragment(thirdPartyRowFragment, props.thirdPartyKey);
  const [deleteThirdParty] = useMutation<ThirdPartyRowDeleteMutation>(
    deleteThirdPartyMutation,
  );
  const confirm = useConfirm();
  const { toast } = useToast();

  const latestAssessment = thirdParty.riskAssessments?.edges[0]?.node;
  const thirdPartyUrl = `/organizations/${organizationId}/third-parties/${thirdParty.id}/overview`;

  const onDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteThirdParty({
            variables: {
              input: { thirdPartyId: thirdParty.id },
              connections: [props.connectionId],
            },
            onCompleted() {
              resolve();
            },
            onError(error) {
              toast({
                title: t("thirdPartyRow.messages.error"),
                description: formatError(
                  t("thirdPartyRow.errors.delete"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("thirdPartyRow.deleteConfirmation", {
          name: thirdParty.name || t("thirdPartyRow.unnamed"),
        }),
      },
    );
  };

  return (
    <Tr to={thirdPartyUrl}>
      <Td>
        <div className="flex gap-2 items-center">
          <Avatar name={thirdParty.name} src={faviconUrl(thirdParty.websiteUrl)} />
          <div>{thirdParty.name}</div>
        </div>
      </Td>
      <Td>
        {latestAssessment?.createdAt
          ? dateFormat(i18n.language, latestAssessment.createdAt)
          : t("thirdPartyRow.notAssessed")}
      </Td>
      <Td>
        <RiskBadge level={latestAssessment?.dataSensitivity ?? "NONE"} />
      </Td>
      <Td>
        <RiskBadge level={latestAssessment?.businessImpact ?? "NONE"} />
      </Td>
      {props.hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {thirdParty.canDelete && (
              <DropdownItem
                onClick={onDelete}
                variant="danger"
                icon={IconTrashCan}
              >
                {t("thirdPartyRow.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
