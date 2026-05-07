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

import { formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import type { PreloadedQuery } from "react-relay";
import { graphql, useMutation, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { AccessReviewCampaignsTabDeleteMutation } from "#/__generated__/core/AccessReviewCampaignsTabDeleteMutation.graphql";
import type { AccessReviewCampaignsTabFragment$key } from "#/__generated__/core/AccessReviewCampaignsTabFragment.graphql";
import type { AccessReviewCampaignsTabPaginationQuery } from "#/__generated__/core/AccessReviewCampaignsTabPaginationQuery.graphql";
import type { AccessReviewCampaignsTabQuery } from "#/__generated__/core/AccessReviewCampaignsTabQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { statusBadgeVariant, statusLabel } from "../_components/accessReviewHelpers";
import { CreateAccessReviewCampaignDialog } from "../dialogs/CreateAccessReviewCampaignDialog";

export const accessReviewCampaignsTabQuery = graphql`
  query AccessReviewCampaignsTabQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateCampaign: permission(action: "core:access-review-campaign:create")
        ...AccessReviewCampaignsTabFragment
      }
    }
  }
`;

const campaignsFragment = graphql`
  fragment AccessReviewCampaignsTabFragment on Organization
  @refetchable(queryName: "AccessReviewCampaignsTabPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: {
      type: "AccessReviewCampaignOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    accessReviewCampaigns(
      first: $first
      after: $after
      orderBy: $order
    ) @connection(key: "AccessReviewCampaignsTab_accessReviewCampaigns") {
      __id
      edges {
        node {
          id
          name
          status
          createdAt
          canDelete: permission(action: "core:access-review-campaign:delete")
        }
      }
    }
  }
`;

const deleteCampaignMutation = graphql`
  mutation AccessReviewCampaignsTabDeleteMutation(
    $input: DeleteAccessReviewCampaignInput!
    $connections: [ID!]!
  ) {
    deleteAccessReviewCampaign(input: $input) {
      deletedAccessReviewCampaignId @deleteEdge(connections: $connections)
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<AccessReviewCampaignsTabQuery>;
};

export default function AccessReviewCampaignsTab({ queryRef }: Props) {
  const { __, dateFormat } = useTranslate();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const { toast } = useToast();

  const { organization } = usePreloadedQuery(accessReviewCampaignsTabQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const {
    data: { accessReviewCampaigns },
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    AccessReviewCampaignsTabPaginationQuery,
    AccessReviewCampaignsTabFragment$key
  >(campaignsFragment, organization);

  const [deleteCampaign] = useMutation<AccessReviewCampaignsTabDeleteMutation>(
    deleteCampaignMutation,
  );

  // Only DRAFT and CANCELLED campaigns can be deleted (enforced by the backend).
  const isDeletableStatus = (status: string) =>
    status === "DRAFT" || status === "CANCELLED";

  const handleDelete = (campaignId: string, campaignName: string) => {
    confirm(
      () => {
        deleteCampaign({
          variables: {
            input: { accessReviewCampaignId: campaignId },
            connections: [accessReviewCampaigns.__id],
          },
          onCompleted(_, errors) {
            if (errors?.length) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete campaign"),
                  errors as GraphQLError[],
                ),
                variant: "error",
              });
              return;
            }
            toast({
              title: __("Success"),
              description: __("Campaign deleted successfully."),
              variant: "success",
            });
          },
          onError(error) {
            toast({
              title: __("Error"),
              description: formatError(
                __("Failed to delete campaign"),
                error as GraphQLError,
              ),
              variant: "error",
            });
          },
        });
      },
      {
        message: sprintf(
          __("This will permanently delete \"%s\". This action cannot be undone."),
          campaignName,
        ),
      },
    );
  };

  const hasActions = accessReviewCampaigns.edges.some(
    edge => edge.node.canDelete && isDeletableStatus(edge.node.status),
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        {organization.canCreateCampaign && (
          <CreateAccessReviewCampaignDialog
            organizationId={organizationId}
            connectionId={accessReviewCampaigns.__id}
          >
            <Button icon={IconPlusLarge}>
              {__("New campaign")}
            </Button>
          </CreateAccessReviewCampaignDialog>
        )}
      </div>

      {accessReviewCampaigns.edges.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Name")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Created at")}</Th>
                    {hasActions && <Th className="w-12"></Th>}
                  </Tr>
                </Thead>
                <Tbody>
                  {accessReviewCampaigns.edges.map((edge) => {
                    const canDeleteRow
                      = edge.node.canDelete && isDeletableStatus(edge.node.status);
                    return (
                      <Tr
                        key={edge.node.id}
                        to={`/organizations/${organizationId}/access-reviews/campaigns/${edge.node.id}`}
                      >
                        <Td>{edge.node.name}</Td>
                        <Td>
                          <Badge variant={statusBadgeVariant(edge.node.status)}>
                            {statusLabel(__, edge.node.status)}
                          </Badge>
                        </Td>
                        <Td>
                          {dateFormat(edge.node.createdAt)}
                        </Td>
                        {hasActions && (
                          <Td noLink width={50} className="text-end">
                            {canDeleteRow && (
                              <ActionDropdown>
                                <DropdownItem
                                  icon={IconTrashCan}
                                  variant="danger"
                                  onSelect={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    handleDelete(edge.node.id, edge.node.name);
                                  }}
                                >
                                  {__("Delete")}
                                </DropdownItem>
                              </ActionDropdown>
                            )}
                          </Td>
                        )}
                      </Tr>
                    );
                  })}
                </Tbody>
              </Table>

              {hasNext && (
                <div className="p-4 border-t">
                  <Button
                    variant="secondary"
                    onClick={() => loadNext(20)}
                    disabled={isLoadingNext}
                  >
                    {isLoadingNext
                      ? __("Loading...")
                      : __("Load more")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-8">
                <p className="text-txt-tertiary">
                  {__("No access review campaigns yet. Create your first campaign to start reviewing access.")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
