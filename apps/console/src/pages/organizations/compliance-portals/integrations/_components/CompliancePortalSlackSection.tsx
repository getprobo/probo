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

import { WarningCircleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { ComboboxPaged } from "@probo/ui/src/v2/Combobox/ComboboxPaged";
import { Field } from "@probo/ui/src/v2/form/Field";
import { SlackLogo } from "@probo/ui/src/v2/SlackLogo/SlackLogo";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Trans, useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalSlackSection_compliancePortal$key } from "#/__generated__/core/CompliancePortalSlackSection_compliancePortal.graphql";
import type { CompliancePortalSlackSectionClearMutation } from "#/__generated__/core/CompliancePortalSlackSectionClearMutation.graphql";
import type { CompliancePortalSlackSectionSetMutation } from "#/__generated__/core/CompliancePortalSlackSectionSetMutation.graphql";
import { tonedCard } from "#/components/TonedCard/variants";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { useSlackChannelPage } from "../_lib/useSlackChannelPage";
import { slackSection } from "../variants";

const fragment = graphql`
  fragment CompliancePortalSlackSection_compliancePortal on CompliancePortal {
    id
    canConfigureSlack: permission(action: "core:connector:initiate")
    slackbotNotificationChannel {
      channelId
      channelName
    }
    organization {
      canInstallSlackbot: permission(action: "core:connector:initiate")
      slackbotAvailable
      slackbotInstallation {
        active
      }
    }
  }
`;

const setMutation = graphql`
  mutation CompliancePortalSlackSectionSetMutation(
    $input: SetSlackbotNotificationChannelInput!
  ) {
    setSlackbotNotificationChannel(input: $input) {
      channel {
        channelId
        channelName
      }
    }
  }
`;

const clearMutation = graphql`
  mutation CompliancePortalSlackSectionClearMutation(
    $input: ClearSlackbotNotificationChannelInput!
  ) {
    clearSlackbotNotificationChannel(input: $input) {
      channelId
    }
  }
`;

interface CompliancePortalSlackSectionProps {
  compliancePortalKey: CompliancePortalSlackSection_compliancePortal$key;
}

export function CompliancePortalSlackSection({
  compliancePortalKey,
}: CompliancePortalSlackSectionProps) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");
  const compliancePortal = useFragment(fragment, compliancePortalKey);
  const [setChannel, isSettingChannel]
    = useMutation<CompliancePortalSlackSectionSetMutation>(setMutation);
  const [clearChannel, isClearingChannel]
    = useMutation<CompliancePortalSlackSectionClearMutation>(clearMutation);

  const setNotificationChannel = (channelId: string) => {
    void setChannel({
      variables: {
        input: {
          compliancePortalId: compliancePortal.id,
          channelId,
        },
      },
      updater(store) {
        const channel = store
          .getRootField("setSlackbotNotificationChannel")
          ?.getLinkedRecord("channel");
        const portal = store.get(compliancePortal.id);
        if (channel && portal) {
          portal.setLinkedRecord(channel, "slackbotNotificationChannel");
        }
      },
    });
  };

  const clearNotificationChannel = () => {
    void clearChannel({
      variables: {
        input: { compliancePortalId: compliancePortal.id },
      },
      updater(store) {
        store
          .get(compliancePortal.id)
          ?.setValue(null, "slackbotNotificationChannel");
      },
    });
  };

  const isInstalled = compliancePortal.organization.slackbotInstallation?.active;
  const configuredChannel = compliancePortal.slackbotNotificationChannel;
  const showInstall = !isInstalled
    && compliancePortal.organization.canInstallSlackbot
    && compliancePortal.organization.slackbotAvailable;
  const showChannelConfig = isInstalled === true && compliancePortal.canConfigureSlack;
  const channelPage = useSlackChannelPage(organizationId, showChannelConfig);
  const channels = configuredChannel
    && !channelPage.channels.some(channel => channel.id === configuredChannel.channelId)
    ? [
        {
          id: configuredChannel.channelId,
          name: configuredChannel.channelName,
        },
        ...channelPage.channels,
      ]
    : channelPage.channels;
  const slackListEmpty = channelPage.channels.length === 0 && !channelPage.hasNext;
  const showEmptyCallout = showChannelConfig && slackListEmpty && !channelPage.failed;

  const {
    root, intro, grid, card, lead, copy, channel, channelRow, channelField,
    empty, emptyCopy, emptyCallout,
  } = slackSection();
  const { frame, header, wash, fade, icon: iconSlot, control, body } = tonedCard({
    tone: "sand",
  });

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("slackSection.title")}
        </Heading>
      </div>
      <div className={grid()}>
        <Card variant="ghost" size={2} padding="none" className={frame({ className: card() })}>
          <div className={header()}>
            <div className={wash()} />
            <div className={fade()} />
            <div className={lead()}>
              <div className={iconSlot()}>
                <SlackLogo aria-hidden />
              </div>
              <div className={copy()}>
                <Text size={3} weight="medium" highContrast>
                  {isInstalled
                    ? t("slackSection.connectionTitle")
                    : t("slackSection.notInstalledTitle")}
                </Text>
                <Text size={2} color="neutral">
                  {isInstalled
                    ? t("slackSection.description")
                    : t("slackSection.notInstalled")}
                </Text>
              </div>
            </div>
            {showInstall && (
              <div className={control()}>
                <ButtonAnchor
                  href={getSlackInstallUrl(organizationId)}
                  variant="solid"
                  color="neutral"
                  highContrast
                >
                  {t("slackSection.actions.install")}
                </ButtonAnchor>
              </div>
            )}
          </div>
          {isInstalled && (
            <div className={body()}>
              <div className={empty()}>
                <Callout
                  color="amber"
                  icon={<WarningCircleIcon weight="fill" />}
                  className={emptyCallout()}
                >
                  <div className={emptyCopy()}>
                    <Text size={2} weight="medium" color="current" highContrast>
                      {t("slackSection.loginRequired.title")}
                    </Text>
                    <Text size={2} color="current">
                      <Trans
                        ns="organizations/compliance-portals"
                        i18nKey="slackSection.loginRequired.description"
                        components={{ cmd: <Code size={2} /> }}
                      />
                    </Text>
                  </div>
                </Callout>
              </div>
              {showEmptyCallout && (
                <div className={empty()}>
                  <Callout
                    variant="surface"
                    color="neutral"
                    className={emptyCallout()}
                    icon={<Spinner size={3} />}
                  >
                    <div className={emptyCopy()}>
                      <Text size={2} weight="medium" highContrast>
                        {t("slackSection.channel.emptyTitle")}
                      </Text>
                      <Text size={2}>
                        {t("slackSection.channel.emptyDescription")}
                      </Text>
                    </div>
                  </Callout>
                </div>
              )}
              {showChannelConfig && (
                <div className={channel()}>
                  <div className={channelRow()}>
                    <Field
                      label={t("slackSection.channel.label")}
                      className={channelField()}
                    >
                      <ComboboxPaged
                        items={channels}
                        loading={channelPage.loading}
                        failed={channelPage.failed}
                        searchable={false}
                        hasNext={channelPage.hasNext}
                        onLoadMore={channelPage.onLoadMore}
                        selectedId={configuredChannel?.channelId ?? null}
                        selectedLabel={
                          configuredChannel
                            ? `#${configuredChannel.channelName}`
                            : t("slackSection.channel.placeholder")
                        }
                        ariaLabel={t("slackSection.channel.label")}
                        placeholder={t("slackSection.channel.placeholder")}
                        emptyLabel={t("slackSection.channel.emptyTitle")}
                        errorLabel={t("slackSection.channel.loadError")}
                        loadMoreLabel={t("slackSection.channel.loadMore")}
                        disabled={isSettingChannel || isClearingChannel}
                        onSelect={listedChannel => setNotificationChannel(listedChannel.id)}
                        itemId={listedChannel => listedChannel.id}
                        itemLabel={listedChannel => `#${listedChannel.name}`}
                      />
                    </Field>
                    {configuredChannel && (
                      <Button
                        variant="surface"
                        color="neutral"
                        loading={isClearingChannel}
                        onClick={clearNotificationChannel}
                      >
                        {t("slackSection.actions.clear")}
                      </Button>
                    )}
                  </div>
                  <Text size={1} color="faint">
                    {t("slackSection.channel.help")}
                  </Text>
                </div>
              )}
            </div>
          )}
        </Card>
      </div>
    </section>
  );
}

function getSlackInstallUrl(organizationId: string): string {
  const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/slackbot/install/initiate", baseUrl);
  url.searchParams.set("organization_id", organizationId);
  url.searchParams.set("continue", window.location.pathname);
  return url.toString();
}
