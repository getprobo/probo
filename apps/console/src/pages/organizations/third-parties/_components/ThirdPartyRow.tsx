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

import { faviconUrl, formatDate, formatError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
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
                title: __("Error"),
                description: formatError(
                  __("Failed to delete third party"),
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
            "This will permanently delete the third party \"%s\". This action cannot be undone.",
          ),
          thirdParty.name || __("Unnamed third party"),
        ),
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
          ? formatDate(latestAssessment.createdAt)
          : __("Not assessed")}
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
                {__("Delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
