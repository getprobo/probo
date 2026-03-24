import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Card,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { graphql, usePaginationFragment } from "react-relay";
import { useOutletContext } from "react-router";

import type { AccessReviewCampaignsTabFragment$key } from "#/__generated__/core/AccessReviewCampaignsTabFragment.graphql";
import type { AccessReviewCampaignsTabPaginationQuery } from "#/__generated__/core/AccessReviewCampaignsTabPaginationQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

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
      edges {
        node {
          id
          name
          status
          createdAt
          startedAt
          completedAt
        }
      }
    }
  }
`;

function statusBadgeVariant(status: string): "neutral" | "info" | "warning" | "success" | "danger" {
  switch (status) {
    case "DRAFT":
      return "neutral";
    case "IN_PROGRESS":
      return "info";
    case "PENDING_ACTIONS":
      return "warning";
    case "COMPLETED":
      return "success";
    case "FAILED":
    case "CANCELLED":
      return "danger";
    default:
      return "neutral";
  }
}

function formatStatus(status: string): string {
  return status.replace(/_/g, " ");
}

export default function AccessReviewCampaignsTab() {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { organizationRef } = useOutletContext<{ organizationRef: AccessReviewCampaignsTabFragment$key }>();

  const {
    data: { accessReviewCampaigns },
  } = usePaginationFragment<
    AccessReviewCampaignsTabPaginationQuery,
    AccessReviewCampaignsTabFragment$key
  >(campaignsFragment, organizationRef);

  if (!accessReviewCampaigns || accessReviewCampaigns.edges.length === 0) {
    return (
      <Card padded>
        <div className="text-center py-8">
          <p className="text-txt-tertiary">
            {__("No access review campaigns yet. Create your first campaign to start reviewing access.")}
          </p>
        </div>
      </Card>
    );
  }

  return (
    <Card>
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Status")}</Th>
            <Th>{__("Created at")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {accessReviewCampaigns.edges.map((edge) => (
            <Tr
              key={edge.node.id}
              to={`/organizations/${organizationId}/access-reviews/campaigns/${edge.node.id}`}
            >
              <Td>{edge.node.name}</Td>
              <Td>
                <Badge variant={statusBadgeVariant(edge.node.status)}>
                  {formatStatus(edge.node.status)}
                </Badge>
              </Td>
              <Td>
                {new Date(edge.node.createdAt as string).toLocaleDateString()}
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </Card>
  );
}
