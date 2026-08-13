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

import { Button, Card, Slack, useConfirm } from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { SlackBotSettingsPage_organization$key } from "#/__generated__/core/SlackBotSettingsPage_organization.graphql";
import type { SlackBotSettingsPageQuery } from "#/__generated__/core/SlackBotSettingsPageQuery.graphql";
import type { SlackBotSettingsPageUninstallMutation } from "#/__generated__/core/SlackBotSettingsPageUninstallMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

export const slackBotSettingsPageQuery = graphql`
  query SlackBotSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...SlackBotSettingsPage_organization
      }
    }
  }
`;

const slackBotSettingsPageFragment = graphql`
  fragment SlackBotSettingsPage_organization on Organization {
    canConnectSlack: permission(action: "core:connector:initiate")
    canUninstallSlack: permission(action: "core:connector:delete")
    slackbotAvailable
    slackbotInstallation {
      active
    }
  }
`;

const uninstallMutation = graphql`
  mutation SlackBotSettingsPageUninstallMutation(
    $input: UninstallSlackbotInput!
  ) {
    uninstallSlackbot(input: $input) {
      organization {
        slackbotInstallation {
          active
        }
      }
    }
  }
`;

interface SlackBotSettingsPageProps {
  queryRef: PreloadedQuery<SlackBotSettingsPageQuery>;
}

export function SlackBotSettingsPage({ queryRef }: SlackBotSettingsPageProps) {
  const data = usePreloadedQuery<SlackBotSettingsPageQuery>(
    slackBotSettingsPageQuery,
    queryRef,
  );

  if (data.organization.__typename !== "Organization") {
    throw new Error("Relay node is not an organization");
  }

  return (
    <SlackBotSettings organizationKey={data.organization} />
  );
}

function SlackBotSettings({
  organizationKey,
}: {
  organizationKey: SlackBotSettingsPage_organization$key;
}) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const confirm = useConfirm();
  const organization = useFragment(
    slackBotSettingsPageFragment,
    organizationKey,
  );
  const [uninstallSlackbot, isUninstalling]
    = useMutation<SlackBotSettingsPageUninstallMutation>(
      uninstallMutation,
      {
        successMessage: t("slackBotSettingsPage.messages.uninstalled"),
        errorToast: t("slackBotSettingsPage.errors.uninstall"),
      },
    );

  const uninstall = () => {
    confirm(
      async () => {
        await uninstallSlackbot({
          variables: {
            input: { organizationId },
          },
        });
      },
      { message: t("slackBotSettingsPage.uninstallConfirmation") },
    );
  };

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">
        {t("slackBotSettingsPage.title")}
      </h2>
      <Card padded className="space-y-4">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 flex items-center justify-center bg-subtle rounded">
            <Slack className="h-6 w-6" />
          </div>
          <div className="mr-auto">
            <h3 className="text-base font-semibold">
              {t("slackBotSettingsPage.connectionTitle")}
            </h3>
            <p className="text-sm text-txt-tertiary">
              {organization.slackbotInstallation?.active
                ? t("slackBotSettingsPage.installed")
                : t("slackBotSettingsPage.emptyDescription")}
            </p>
          </div>
          <div className="flex items-center gap-2">
            {organization.slackbotInstallation
              && organization.canUninstallSlack && (
              <Button
                variant="danger"
                onClick={uninstall}
                disabled={isUninstalling}
              >
                {isUninstalling
                  ? t("slackBotSettingsPage.actions.uninstalling")
                  : t("slackBotSettingsPage.actions.uninstall")}
              </Button>
            )}
            {organization.canConnectSlack && organization.slackbotAvailable && (
              <Button variant="secondary" asChild>
                <a href={getSlackInstallUrl(organizationId)}>
                  {organization.slackbotInstallation
                    ? t("slackBotSettingsPage.actions.configure")
                    : t("slackBotSettingsPage.actions.install")}
                </a>
              </Button>
            )}
          </div>
        </div>
      </Card>
    </div>
  );
}

function getSlackInstallUrl(organizationId: string): string {
  const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/slackbot/install/initiate", baseUrl);
  url.searchParams.append("organization_id", organizationId);
  url.searchParams.append(
    "continue",
    `/organizations/${organizationId}/settings/slackbot`,
  );
  return url.toString();
}
