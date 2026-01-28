import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Card } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageSlackSectionFragment$key } from "#/__generated__/core/CompliancePageSlackSectionFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment CompliancePageSlackSectionFragment on Organization {
    compliancePage: trustCenter {
      canUpdate: permission(action: "core:trust-center:update")
    }
    slackConnections(first: 100) {
      edges {
        node {
          id
          channel
          createdAt
        }
      }
    }
  }
`;

export function CompliancePageSlackSection(props: { fragmentRef: CompliancePageSlackSectionFragment$key }) {
  const { fragmentRef } = props;

  const organizationId = useOrganizationId();
  const { __, dateTimeFormat } = useTranslate();

  const organization = useFragment<CompliancePageSlackSectionFragment$key>(fragment, fragmentRef);

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{__("Integrations")}</h2>
      <Card padded>
        <div className="space-y-2">
          {organization.slackConnections.edges.map(({ node: slackConnection }) => (
            <Card
              key={slackConnection.id}
              padded
              className="flex items-center gap-3"
            >
              <div>
                <img src={`/${slackConnection.id.toLowerCase()}.png`} alt="" />
              </div>
              <div className="mr-auto">
                <h3 className="text-base font-semibold">Slack</h3>
                <p className="text-sm text-txt-tertiary">
                  {slackConnection.createdAt
                    ? (
                        <>
                          {sprintf(
                            __("Connected on %s"),
                            dateTimeFormat(slackConnection.createdAt),
                          )}
                          {slackConnection.channel && (
                            <>
                              {" • "}
                              {sprintf(__("Channel: %s"), slackConnection.channel)}
                            </>
                          )}
                        </>
                      )
                    : __("Manage your trust center access with slack")}
                </p>
              </div>
              {slackConnection.createdAt
                ? (
                    <div>
                      <Badge variant="success" size="md">
                        {__("Connected")}
                      </Badge>
                    </div>
                  )
                : (
                    organization.compliancePage?.canUpdate && (
                      <Button variant="secondary" asChild>
                        <a href={getSlackConnectionUrl(organizationId)}>{__("Connect")}</a>
                      </Button>
                    )
                  )}
            </Card>
          ))}
        </div>
      </Card>
    </div>
  );
}

function getSlackConnectionUrl(organizationId: string): string {
  const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/connectors/initiate", baseUrl);
  url.searchParams.append("organization_id", organizationId);
  url.searchParams.append("provider", "SLACK");
  const redirectUrl = `/organizations/${organizationId}/compliance-page`;
  url.searchParams.append("continue", redirectUrl);
  const finalUrl = url.toString();
  return finalUrl;
};
