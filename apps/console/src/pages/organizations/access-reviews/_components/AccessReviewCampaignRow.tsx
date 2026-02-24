import { formatDate, formatError, type GraphQLError, promisifyMutation, sprintf } from "@probo/helpers";
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

import type { AccessReviewCampaignRowDeleteMutation } from "#/__generated__/core/AccessReviewCampaignRowDeleteMutation.graphql";
import type { AccessReviewCampaignRowFragment$key } from "#/__generated__/core/AccessReviewCampaignRowFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment AccessReviewCampaignRowFragment on AccessReviewCampaign {
    id
    name
    status
    createdAt
    canDelete: permission(action: "core:access-review-campaign:delete")
  }
`;

const deleteMutation = graphql`
  mutation AccessReviewCampaignRowDeleteMutation(
    $input: DeleteAccessReviewCampaignInput!
    $connections: [ID!]!
  ) {
    deleteAccessReviewCampaign(input: $input) {
      deletedAccessReviewCampaignId @deleteEdge(connections: $connections)
    }
  }
`;

function statusBadgeVariant(status: string) {
  switch (status) {
    case "DRAFT":
      return "neutral" as const;
    case "IN_PROGRESS":
      return "info" as const;
    case "PENDING_ACTIONS":
      return "warning" as const;
    case "COMPLETED":
      return "success" as const;
    case "CANCELLED":
      return "danger" as const;
    default:
      return "neutral" as const;
  }
}

type Props = {
  fKey: AccessReviewCampaignRowFragment$key;
  connectionId: string;
};

export function AccessReviewCampaignRow({ fKey, connectionId }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const { toast } = useToast();

  const campaign = useFragment(fragment, fKey);
  const canDelete = campaign.canDelete;

  const [deleteCampaign] = useMutation<AccessReviewCampaignRowDeleteMutation>(deleteMutation);

  const handleDelete = () => {
    if (!campaign.id || !campaign.name) {
      return alert(__("Failed to delete campaign: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(deleteCampaign)({
          variables: {
            input: {
              accessReviewCampaignId: campaign.id,
            },
            connections: [connectionId],
          },
        }).catch((error) => {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to delete campaign"),
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
          campaign.name,
        ),
      },
    );
  };

  const detailUrl = `/organizations/${organizationId}/access-reviews/campaigns/${campaign.id}`;

  return (
    <Tr to={detailUrl}>
      <Td>{campaign.name}</Td>
      <Td>
        <Badge variant={statusBadgeVariant(campaign.status!)} size="sm">
          {campaign.status}
        </Badge>
      </Td>
      <Td>
        <time dateTime={campaign.createdAt}>
          {formatDate(campaign.createdAt)}
        </time>
      </Td>
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
