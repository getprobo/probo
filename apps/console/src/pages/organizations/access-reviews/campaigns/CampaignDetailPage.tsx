import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Breadcrumb,
  Card,
  IconChevronDown,
  IconChevronRight,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CampaignDetailPageQuery } from "#/__generated__/core/CampaignDetailPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const campaignDetailPageQuery = graphql`
  query CampaignDetailPageQuery($campaignId: ID!) {
    node(id: $campaignId) {
      __typename
      ... on AccessReviewCampaign {
        id
        name
        status
        createdAt
        startedAt
        completedAt
        scopeSources {
          id
          name
          fetchStatus
          fetchedAccountsCount
          entries(first: 50) {
            edges {
              node {
                id
                email
                fullName
                role
                isAdmin
                mfaStatus
                lastLogin
                decision
                flag
              }
            }
            pageInfo {
              hasNextPage
            }
          }
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

function decisionBadgeVariant(decision: string): "neutral" | "success" | "danger" {
  switch (decision) {
    case "APPROVED":
      return "success";
    case "REVOKED":
      return "danger";
    default:
      return "neutral";
  }
}

function flagBadgeVariant(flag: string): "neutral" | "warning" | "danger" {
  switch (flag) {
    case "RISK":
      return "danger";
    case "REVIEW":
      return "warning";
    default:
      return "neutral";
  }
}

type Props = {
  queryRef: PreloadedQuery<CampaignDetailPageQuery>;
};

export default function CampaignDetailPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(campaignDetailPageQuery, queryRef);

  if (data.node.__typename !== "AccessReviewCampaign") {
    throw new Error("Campaign not found");
  }

  const campaign = data.node;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Access Reviews"),
            to: `/organizations/${organizationId}/access-reviews`,
          },
          { label: campaign.name },
        ]}
      />

      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-semibold">{campaign.name}</h1>
        <Badge variant={statusBadgeVariant(campaign.status)}>
          {formatStatus(campaign.status)}
        </Badge>
      </div>

      <div className="space-y-4">
        {campaign.scopeSources.map((source) => (
          <ScopeSourceCard key={source.id} source={source} />
        ))}

        {campaign.scopeSources.length === 0 && (
          <Card padded>
            <div className="text-center py-8">
              <p className="text-txt-tertiary">
                {__("No sources configured for this campaign.")}
              </p>
            </div>
          </Card>
        )}
      </div>
    </div>
  );
}

type ScopeSource = NonNullable<
  Extract<
    CampaignDetailPageQuery["response"]["node"],
    { readonly __typename: "AccessReviewCampaign" }
  >["scopeSources"]
>[number];

function ScopeSourceCard({ source }: { source: ScopeSource }) {
  const { __ } = useTranslate();
  const [expanded, setExpanded] = useState(false);

  return (
    <Card>
      <button
        type="button"
        className="flex w-full items-center justify-between p-4 text-left hover:bg-bg-subtle transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-3">
          {expanded
            ? <IconChevronDown className="size-4 text-txt-tertiary" />
            : <IconChevronRight className="size-4 text-txt-tertiary" />}
          <span className="font-medium">{source.name}</span>
          <Badge variant="neutral">
            {source.fetchedAccountsCount} {__("accounts")}
          </Badge>
          <Badge variant={source.fetchStatus === "SUCCESS" ? "success" : "info"}>
            {formatStatus(source.fetchStatus)}
          </Badge>
        </div>
      </button>

      {expanded && (
        <div className="border-t">
          {source.entries.edges.length === 0
            ? (
                <div className="px-4 py-6 text-center text-txt-tertiary">
                  {__("No entries found for this source.")}
                </div>
              )
            : (
                <Table>
                  <Thead>
                    <Tr>
                      <Th>{__("Name")}</Th>
                      <Th>{__("Email")}</Th>
                      <Th>{__("Role")}</Th>
                      <Th>{__("Admin")}</Th>
                      <Th>{__("MFA")}</Th>
                      <Th>{__("Last login")}</Th>
                      <Th>{__("Flag")}</Th>
                      <Th>{__("Decision")}</Th>
                    </Tr>
                  </Thead>
                  <Tbody>
                    {source.entries.edges.map((edge) => (
                      <Tr key={edge.node.id}>
                        <Td>{edge.node.fullName}</Td>
                        <Td>{edge.node.email}</Td>
                        <Td>{edge.node.role}</Td>
                        <Td>{edge.node.isAdmin ? __("Yes") : __("No")}</Td>
                        <Td>
                          <Badge variant={edge.node.mfaStatus === "ENABLED" ? "success" : "neutral"}>
                            {formatStatus(edge.node.mfaStatus)}
                          </Badge>
                        </Td>
                        <Td>
                          {edge.node.lastLogin
                            ? formatDate(edge.node.lastLogin as string)
                            : "—"}
                        </Td>
                        <Td>
                          {edge.node.flag !== "NONE" && (
                            <Badge variant={flagBadgeVariant(edge.node.flag)}>
                              {edge.node.flag}
                            </Badge>
                          )}
                        </Td>
                        <Td>
                          {edge.node.decision !== "PENDING" && (
                            <Badge variant={decisionBadgeVariant(edge.node.decision)}>
                              {edge.node.decision}
                            </Badge>
                          )}
                        </Td>
                      </Tr>
                    ))}
                  </Tbody>
                </Table>
              )}

          {source.entries.pageInfo.hasNextPage && (
            <div className="p-4 border-t text-center">
              <p className="text-sm text-txt-tertiary">
                {__("Showing first 50 entries. More entries available.")}
              </p>
            </div>
          )}
        </div>
      )}
    </Card>
  );
}
