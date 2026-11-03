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

import { useConfirm } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { SlackLogo } from "@probo/ui/src/v2/SlackLogo/SlackLogo";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
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

import { slackBotSettingsPage } from "./variants";

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
      teamName
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
          teamName
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

  const installation = organization.slackbotInstallation;
  const active = installation?.active === true;
  const { root, grid, action } = slackBotSettingsPage();
  const showConnect = organization.canConnectSlack && organization.slackbotAvailable;
  const showUninstall = installation != null && organization.canUninstallSlack;

  return (
    <section className={root()}>
      <Heading level={2} size={4} weight="medium" highContrast>
        {t("slackBotSettingsPage.title")}
      </Heading>
      <div className={grid()}>
        <SlackBotCard
          connected={active}
          title={t("slackBotSettingsPage.connectionTitle")}
          description={active
            ? (installation?.teamName
                ? t("slackBotSettingsPage.installedIn", { workspace: installation.teamName })
                : t("slackBotSettingsPage.installed"))
            : t("slackBotSettingsPage.emptyDescription")}
          actions={(
            <>
              {showConnect && (
                <ButtonAnchor
                  href={getSlackInstallUrl(organizationId)}
                  size={2}
                  variant="solid"
                  color="neutral"
                  highContrast
                  className={action()}
                >
                  {installation
                    ? t("slackBotSettingsPage.actions.configure")
                    : t("slackBotSettingsPage.actions.install")}
                </ButtonAnchor>
              )}
              {showUninstall && (
                <Button
                  size={2}
                  variant="solid"
                  color="red"
                  className={action()}
                  loading={isUninstalling}
                  onClick={uninstall}
                >
                  {isUninstalling
                    ? t("slackBotSettingsPage.actions.uninstalling")
                    : t("slackBotSettingsPage.actions.uninstall")}
                </Button>
              )}
            </>
          )}
        />
      </div>
    </section>
  );
}

function SlackBotCard({
  connected,
  title,
  description,
  actions,
}: {
  connected: boolean;
  title: string;
  description: string;
  actions: ReactNode;
}) {
  const { frame, wash, header, icon, copy, description: descriptionSlot, body }
    = slackBotSettingsPage({ connected });

  return (
    <Card variant="soft" padding="none" size={3} className={frame()}>
      <div className={wash()} />
      <div className={header()}>
        <div className={icon()}>
          <SlackLogo aria-hidden />
        </div>
        <div className={copy()}>
          <Heading level={2} size={2} weight="medium" highContrast>
            {title}
          </Heading>
          <Text size={1} color="current" className={descriptionSlot()}>
            {description}
          </Text>
        </div>
      </div>
      <div className={body()}>
        {actions}
      </div>
    </Card>
  );
}

function getSlackInstallUrl(organizationId: string): string {
  const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/slackbot/install/initiate", baseUrl);
  url.searchParams.set("organization_id", organizationId);
  url.searchParams.set(
    "continue",
    `/organizations/${organizationId}/settings/slackbot`,
  );
  return url.toString();
}
