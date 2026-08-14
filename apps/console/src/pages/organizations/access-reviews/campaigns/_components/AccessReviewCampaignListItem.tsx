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
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AccessReviewCampaignListItem_campaign$key } from "#/__generated__/core/AccessReviewCampaignListItem_campaign.graphql";

import {
  isCampaignDeletableStatus,
  statusBadgeVariant,
} from "../../_components/accessReviewHelpers";

const accessReviewCampaignListItemFragment = graphql`
  fragment AccessReviewCampaignListItem_campaign on AccessReviewCampaign {
    id
    name
    status
    createdAt
    canDelete: permission(action: "access-review:campaign:delete")
  }
`;

interface AccessReviewCampaignListItemProps {
  campaignKey: AccessReviewCampaignListItem_campaign$key;
  organizationId: string;
  hasActions: boolean;
  onDelete: (campaignId: string, campaignName: string) => void;
}

export function AccessReviewCampaignListItem({
  campaignKey,
  organizationId,
  hasActions,
  onDelete,
}: AccessReviewCampaignListItemProps) {
  const { i18n, t } = useTranslation();
  const campaign = useFragment(
    accessReviewCampaignListItemFragment,
    campaignKey,
  );
  const canDelete
    = campaign.canDelete && isCampaignDeletableStatus(campaign.status);

  return (
    <Tr
      to={`/organizations/${organizationId}/access-reviews/campaigns/${campaign.id}`}
    >
      <Td>{campaign.name}</Td>
      <Td>
        <Badge variant={statusBadgeVariant(campaign.status)}>
          {t(
            `accessReviewCampaignsPage.status.${campaign.status.toLowerCase()}`,
          )}
        </Badge>
      </Td>
      <Td>{dateFormat(i18n.language, campaign.createdAt)}</Td>
      {hasActions && (
        <Td noLink width={50} className="text-end">
          {canDelete && (
            <ActionDropdown>
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                  onDelete(campaign.id, campaign.name);
                }}
              >
                {t("accessReviewCampaignsPage.actions.delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </Td>
      )}
    </Tr>
  );
}
