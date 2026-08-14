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
import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  IconPlusLarge,
  PageHeader,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, useMutation, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { AccessReviewCampaignsPageDeleteMutation } from "#/__generated__/core/AccessReviewCampaignsPageDeleteMutation.graphql";
import type { AccessReviewCampaignsPageFragment$key } from "#/__generated__/core/AccessReviewCampaignsPageFragment.graphql";
import type { AccessReviewCampaignsPagePaginationQuery } from "#/__generated__/core/AccessReviewCampaignsPagePaginationQuery.graphql";
import type { AccessReviewCampaignsPageQuery } from "#/__generated__/core/AccessReviewCampaignsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { isCampaignDeletableStatus } from "../_components/accessReviewHelpers";
import { CreateAccessReviewCampaignDialog } from "../dialogs/CreateAccessReviewCampaignDialog";

import { AccessReviewCampaignListItem } from "./_components/AccessReviewCampaignListItem";

export const accessReviewCampaignsPageQuery = graphql`
  query AccessReviewCampaignsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateCampaign: permission(action: "access-review:campaign:create")
        ...AccessReviewCampaignsPageFragment
      }
    }
  }
`;

const campaignsFragment = graphql`
  fragment AccessReviewCampaignsPageFragment on Organization
  @refetchable(queryName: "AccessReviewCampaignsPagePaginationQuery")
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
    ) @connection(key: "AccessReviewCampaignsPage_accessReviewCampaigns") {
      __id
      edges {
        node {
          id
          status
          canDelete: permission(action: "access-review:campaign:delete")
          ...AccessReviewCampaignListItem_campaign
        }
      }
    }
  }
`;

const deleteCampaignMutation = graphql`
  mutation AccessReviewCampaignsPageDeleteMutation(
    $input: DeleteAccessReviewCampaignInput!
    $connections: [ID!]!
  ) {
    deleteAccessReviewCampaign(input: $input) {
      deletedAccessReviewCampaignId @deleteEdge(connections: $connections)
    }
  }
`;

interface AccessReviewCampaignsPageProps {
  queryRef: PreloadedQuery<AccessReviewCampaignsPageQuery>;
}

export function AccessReviewCampaignsPage({ queryRef }: AccessReviewCampaignsPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const { toast } = useToast();

  usePageTitle(t("accessReviewCampaignsPage.title"));

  const { organization } = usePreloadedQuery<AccessReviewCampaignsPageQuery>(
    accessReviewCampaignsPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const {
    data: { accessReviewCampaigns },
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    AccessReviewCampaignsPagePaginationQuery,
    AccessReviewCampaignsPageFragment$key
  >(campaignsFragment, organization);

  const [deleteCampaign] = useMutation<AccessReviewCampaignsPageDeleteMutation>(
    deleteCampaignMutation,
  );

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
                title: t("accessReviewCampaignsPage.messages.error"),
                description: formatError(
                  t("accessReviewCampaignsPage.errors.delete"),
                  errors,
                ),
                variant: "error",
              });
              return;
            }
            toast({
              title: t("accessReviewCampaignsPage.messages.success"),
              description: t("accessReviewCampaignsPage.messages.deleted"),
              variant: "success",
            });
          },
          onError(error) {
            toast({
              title: t("accessReviewCampaignsPage.messages.error"),
              description: formatError(
                t("accessReviewCampaignsPage.errors.delete"),
                error,
              ),
              variant: "error",
            });
          },
        });
      },
      {
        message: t("accessReviewCampaignsPage.deleteConfirmation", {
          name: campaignName,
        }),
      },
    );
  };

  const hasActions = accessReviewCampaigns.edges.some(
    edge => edge.node.canDelete && isCampaignDeletableStatus(edge.node.status),
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("accessReviewCampaignsPage.title")}
        description={t("accessReviewCampaignsPage.description")}
      >
        {organization.canCreateCampaign && (
          <CreateAccessReviewCampaignDialog
            organizationId={organizationId}
            connectionId={accessReviewCampaigns.__id}
          >
            <Button icon={IconPlusLarge}>
              {t("accessReviewCampaignsPage.actions.new")}
            </Button>
          </CreateAccessReviewCampaignDialog>
        )}
      </PageHeader>

      {accessReviewCampaigns.edges.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("accessReviewCampaignsPage.columns.name")}</Th>
                    <Th>{t("accessReviewCampaignsPage.columns.status")}</Th>
                    <Th>{t("accessReviewCampaignsPage.columns.createdAt")}</Th>
                    {hasActions && <Th className="w-12"></Th>}
                  </Tr>
                </Thead>
                <Tbody>
                  {accessReviewCampaigns.edges.map(edge => (
                    <AccessReviewCampaignListItem
                      key={edge.node.id}
                      campaignKey={edge.node}
                      organizationId={organizationId}
                      hasActions={hasActions}
                      onDelete={handleDelete}
                    />
                  ))}
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
                      ? t("accessReviewCampaignsPage.actions.loading")
                      : t("accessReviewCampaignsPage.actions.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-8">
                <p className="text-txt-tertiary">
                  {t("accessReviewCampaignsPage.empty")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
