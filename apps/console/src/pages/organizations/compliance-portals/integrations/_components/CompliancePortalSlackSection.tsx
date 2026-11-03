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

import { SpinnerGapIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectList } from "@probo/ui/src/v2/Select/SelectList";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { SlackLogo } from "@probo/ui/src/v2/SlackLogo/SlackLogo";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
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
  const showChannelSelect = channels.length > 0;
  const showStandaloneLoadMore = !showChannelSelect && channelPage.hasNext;

  const channelNames = Object.fromEntries(
    channels.map(listedChannel => [listedChannel.id, `#${listedChannel.name}`]),
  );
  const {
    root, intro, grid, card, lead, copy, channel, channelRow, channelField,
    empty, emptyCopy, emptyCallout, menuList,
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
          {showChannelConfig && (
            <div className={body()}>
              {showEmptyCallout && (
                <div className={empty()}>
                  <Callout
                    color="sky"
                    className={emptyCallout()}
                    icon={<SpinnerGapIcon className="animate-spin" />}
                  >
                    <div className={emptyCopy()}>
                      <Text size={2} weight="medium" color="current" highContrast>
                        {t("slackSection.channel.emptyTitle")}
                      </Text>
                      <Text size={2} color="current">
                        {t("slackSection.channel.emptyDescription")}
                      </Text>
                    </div>
                  </Callout>
                </div>
              )}
              {channelPage.failed && !showChannelSelect && (
                <div className={empty()}>
                  <Callout color="red" className={emptyCallout()}>
                    {t("slackSection.channel.loadError")}
                  </Callout>
                </div>
              )}
              {showStandaloneLoadMore && (
                <div>
                  {channelPage.failedMore && (
                    <Text size={2} color="faint" className="px-3 py-2">
                      {t("slackSection.channel.loadError")}
                    </Text>
                  )}
                  <Button
                    variant="ghost"
                    color="neutral"
                    size={1}
                    loading={channelPage.loadingMore}
                    onClick={channelPage.onLoadMore}
                  >
                    {t("slackSection.channel.loadMore")}
                  </Button>
                </div>
              )}
              {showChannelSelect && (
                <div className={channel()}>
                  <div className={channelRow()}>
                    <Field
                      label={t("slackSection.channel.label")}
                      className={channelField()}
                    >
                      <Select
                        value={configuredChannel?.channelId ?? null}
                        onValueChange={(channelId) => {
                          if (channelId != null) {
                            setNotificationChannel(channelId);
                          }
                        }}
                        disabled={isSettingChannel || isClearingChannel}
                      >
                        <SelectTrigger
                          placeholder={t("slackSection.channel.placeholder")}
                          aria-label={t("slackSection.channel.label")}
                        >
                          {(value: string | null) => (
                            value
                              ? (channelNames[value] ?? `#${value}`)
                              : t("slackSection.channel.placeholder")
                          )}
                        </SelectTrigger>
                        <SelectPopup>
                          <div className={menuList()}>
                            <SelectList>
                              {channels.map(listedChannel => (
                                <SelectItem key={listedChannel.id} value={listedChannel.id}>
                                  {`#${listedChannel.name}`}
                                </SelectItem>
                              ))}
                            </SelectList>
                          </div>
                          {channelPage.hasNext && (
                            <div
                              onMouseDown={(event) => {
                                event.preventDefault();
                              }}
                            >
                              {channelPage.failedMore && (
                                <Text size={2} color="faint" className="px-3 py-2">
                                  {t("slackSection.channel.loadError")}
                                </Text>
                              )}
                              <Button
                                variant="ghost"
                                color="neutral"
                                size={1}
                                className="w-full justify-center"
                                loading={channelPage.loadingMore}
                                onClick={channelPage.onLoadMore}
                              >
                                {t("slackSection.channel.loadMore")}
                              </Button>
                            </div>
                          )}
                        </SelectPopup>
                      </Select>
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
