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

import { formatError } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
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
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, useMutation, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { AccessReviewCampaignsTabDeleteMutation } from "#/__generated__/core/AccessReviewCampaignsTabDeleteMutation.graphql";
import type { AccessReviewCampaignsTabFragment$key } from "#/__generated__/core/AccessReviewCampaignsTabFragment.graphql";
import type { AccessReviewCampaignsTabPaginationQuery } from "#/__generated__/core/AccessReviewCampaignsTabPaginationQuery.graphql";
import type { AccessReviewCampaignsTabQuery } from "#/__generated__/core/AccessReviewCampaignsTabQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { statusBadgeVariant } from "../_components/accessReviewHelpers";
import { CreateAccessReviewCampaignDialog } from "../dialogs/CreateAccessReviewCampaignDialog";

export const accessReviewCampaignsTabQuery = graphql`
  query AccessReviewCampaignsTabQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateCampaign: permission(action: "access-review:campaign:create")
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
          canDelete: permission(action: "access-review:campaign:delete")
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
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const { toast } = useToast();

  const { organization } = usePreloadedQuery<AccessReviewCampaignsTabQuery>(accessReviewCampaignsTabQuery, queryRef);
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
                title: t("accessReviewCampaignsTab.messages.error"),
                description: formatError(
                  t("accessReviewCampaignsTab.errors.delete"),
                  errors,
                ),
                variant: "error",
              });
              return;
            }
            toast({
              title: t("accessReviewCampaignsTab.messages.success"),
              description: t("accessReviewCampaignsTab.messages.deleted"),
              variant: "success",
            });
          },
          onError(error) {
            toast({
              title: t("accessReviewCampaignsTab.messages.error"),
              description: formatError(
                t("accessReviewCampaignsTab.errors.delete"),
                error,
              ),
              variant: "error",
            });
          },
        });
      },
      {
        message: t("accessReviewCampaignsTab.deleteConfirmation", {
          name: campaignName,
        }),
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
              {t("accessReviewCampaignsTab.actions.new")}
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
                    <Th>{t("accessReviewCampaignsTab.columns.name")}</Th>
                    <Th>{t("accessReviewCampaignsTab.columns.status")}</Th>
                    <Th>{t("accessReviewCampaignsTab.columns.createdAt")}</Th>
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
                            {t(
                              `accessReviewCampaignsTab.status.${edge.node.status.toLowerCase()}`,
                            )}
                          </Badge>
                        </Td>
                        <Td>
                          {dateFormat(i18n.language, edge.node.createdAt)}
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
                                  {t("accessReviewCampaignsTab.actions.delete")}
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
                      ? t("accessReviewCampaignsTab.actions.loading")
                      : t("accessReviewCampaignsTab.actions.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-8">
                <p className="text-txt-tertiary">
                  {t("accessReviewCampaignsTab.empty")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
