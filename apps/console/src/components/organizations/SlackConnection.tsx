import { Badge, Button, Card } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { sprintf } from "@probo/helpers";
import type { TrustCenterGraphQuery$data } from "/hooks/graph/__generated__/TrustCenterGraphQuery.graphql";
import { IfAuthorized } from "/permissions/IfAuthorized";

type Props = {
  organizationId: string;
  slackConnections: NonNullable<TrustCenterGraphQuery$data["organization"]["slackConnections"]>["edges"][number]["node"][];
};

export function SlackConnections({ organizationId, slackConnections: connectedSlackConnections }: Props) {
  const { __, dateTimeFormat } = useTranslate();

  const slackConnectionDefinitions = [
    {
      id: "SLACK",
      name: "Slack",
      protocol: "OAUTH2",
      description: __("Manage your trust center access with slack"),
    },
  ];

  const slackConnections = slackConnectionDefinitions.map((def) => {
    const connected = connectedSlackConnections.find((c) => c.id);
    return {
      ...def,
      createdAt: connected?.createdAt,
      channel: connected?.channel,
      channelId: connected?.channelId,
    };
  });

  const getUrl = (provider: string) => {
    const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseUrl);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", provider);
    const trustCenterUrl = `/organizations/${organizationId}/trust-center`;
    url.searchParams.append("continue", trustCenterUrl);
    const finalUrl = url.toString();
    return finalUrl;
  };

  return (
    <div className="space-y-2">
      {slackConnections.map((slackConnection) => (
        <Card key={slackConnection.id} padded className="flex items-center gap-3">
          <div>
            <img src={`/${slackConnection.id.toLowerCase()}.png`} alt="" />
          </div>
          <div className="mr-auto">
            <h3 className="text-base font-semibold">{slackConnection.name}</h3>
            <p className="text-sm text-txt-tertiary">
              {slackConnection.createdAt ? (
                <>
                  {sprintf(
                    __("Connected on %s"),
                    dateTimeFormat(slackConnection.createdAt)
                  )}
                  {slackConnection.channel && (
                    <>
                      {" • "}
                      {sprintf(__("Channel: %s"), slackConnection.channel)}
                    </>
                  )}
                </>
              ) : (
                slackConnection.description
              )}
            </p>
          </div>
          {slackConnection.createdAt ? (
            <div>
              <Badge variant="success" size="md">
                {__("Connected")}
              </Badge>
            </div>
          ) : (
            <IfAuthorized entity="TrustCenter" action="update">
              <Button variant="secondary" asChild>
                <a href={getUrl(slackConnection.id)}>{__("Connect")}</a>
              </Button>
            </IfAuthorized>
          )}
        </Card>
      ))}
    </div>
  );
}
